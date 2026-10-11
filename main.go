package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"co2-exporter/monitor"
	"co2-exporter/prometheus"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
)

var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9110").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	labelTag = kingpin.Flag(
		"label.tag",
		"Tag for exposed metrics.",
	).Short('t').Envar("LABEL_TAG").Default("default").String()
)

// Logger is the application logger.
var Logger *slog.Logger

// InitCli parses command-line options and initializes logging.
func InitCli() {
	logConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, logConfig)
	kingpin.Version(version.Print("co2_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	Logger = promslog.New(logConfig)
}

func pollSensorRecord(ctx context.Context, mon *monitor.DeviceMonitor, exporter *prometheus.Exporter) {
	if err := mon.ReadDataContext(ctx); err != nil {
		if ctx.Err() == nil {
			Logger.Error("Failed to read sensor data", "err", err)
		}
		return
	}
	exporter.SetTemp(mon.GetTemp())
	exporter.SetPpmCo2(mon.GetCo2())
	exporter.SetRh(mon.GetHum())
}

func main() {
	InitCli()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		Logger.Error("Exporter stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	device := monitor.DefaultDevice()
	mon := monitor.NewDeviceMonitor(device)
	mon.SetLogger(Logger)
	disconnect, err := mon.Connect()
	if err != nil {
		return fmt.Errorf("connect to device: %w", err)
	}
	defer disconnect()
	Logger.Debug("monitor device connected")

	// Stop polling before the deferred USB cleanup runs.
	exporter := prometheus.NewExporter()
	exporter.SetLabelTag(*labelTag)

	pollCtx, cancelPoll := context.WithCancel(ctx)
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				pollSensorRecord(pollCtx, mon, exporter)
			}
		}
	}()
	defer func() {
		cancelPoll()
		<-pollDone
	}()

	Logger.Info("Starting Air Co2 Exporter", "version", version.Info())

	// Bind to http service
	handler := http.NewServeMux()
	handler.Handle(*metricsPath, exporter.Handler)
	handler.Handle("/", prometheus.IndexHandler(metricsPath))
	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		Logger.Info("Listening on", "address", *listenAddress)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		Logger.Info("Shutting down Air Co2 Exporter")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		return nil
	}
}
