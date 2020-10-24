package main

import (
	"co2-exporter/monitor"
	"co2-exporter/prometheus"
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// Logger - Default logger
var Logger log.Logger

// InitCli - Config parameters for CLI use
func InitCli() {
	var logConfig = &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, logConfig)
	kingpin.Version(version.Print("co2_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	Logger = promlog.New(logConfig)
}

func pollSensorRecord(mon *monitor.DeviceMonitor, exporter *prometheus.Exporter) {
	err := mon.ReadData()
	if err != nil {
		_ = level.Error(Logger).Log("err", err)
		return
	}
	exporter.SetTemp(mon.GetTemp())
	exporter.SetPpmCo2(mon.GetCo2())
	exporter.SetRh(mon.GetHum())
}

func main() {
	InitCli()
	// Setup device connection
	device := monitor.DefaultDevice()
	mon := monitor.NewDeviceMonitor(device)
	mon.SetLogger(&Logger)
	disconnect, err := mon.Connect()
	if err != nil {
		_ = level.Error(Logger).Log("err", err)
		os.Exit(1)
	}
	_ = level.Debug(Logger).Log("msg", "monitor device connected")

	// Setup prometheus metrics server
	exporter := prometheus.NewExporter()
	exporter.SetLabelTag(*labelTag)

	go func() {
		for {
			time.Sleep(2 * time.Second)
			pollSensorRecord(mon, exporter)
		}
	}()

	_ = level.Info(Logger).Log("msg", "Starting Air Co2 Exporter", "version", version.Info())

	// Bind to http service
	handler := http.NewServeMux()
	handler.Handle(*metricsPath, exporter.Handler)
	handler.Handle("/", prometheus.IndexHandler(metricsPath))
	server := &http.Server{Addr: *listenAddress, Handler: handler}
	go func() {
		_ = level.Info(Logger).Log("msg", "Listening on", "address", *listenAddress)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			_ = level.Error(Logger).Log("err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	signals := make(chan os.Signal, 1)
	shutdown := make(chan bool)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		_ = level.Info(Logger).Log("msg", "Shutting down Air Co2 Exporter")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		disconnect()
		close(shutdown)
	}()

	<-shutdown
	os.Exit(0)
}
