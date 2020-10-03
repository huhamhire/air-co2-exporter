package main

import (
	"co2-exporter/monitor"
	"co2-exporter/prometheus"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
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
)

var Logger log.Logger

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
	exporter.SetDefaultTemp(mon.GetTemp())
	exporter.SetDefaultPpmCo2(mon.GetCo2())
}

func main() {
	InitCli()
	// Setup device connection
	device := monitor.DefaultDevice()
	mon := monitor.NewDeviceMonitor(device)
	mon.SetLogger(&Logger)
	_, err := mon.Connect()
	if err != nil {
		_ = level.Error(Logger).Log("err", err)
		os.Exit(1)
	}
	_ = level.Debug(Logger).Log("msg", "TFA device connected")

	// Setup prometheus metrics server
	exporter := prometheus.NewExporter()

	go func() {
		for {
			time.Sleep(2 * time.Second)
			pollSensorRecord(mon, exporter)
		}
	}()

	_ = level.Info(Logger).Log("msg", "Starting Co2 Exporter", "version", version.Info())

	// Bind to http service
	http.Handle(*metricsPath, exporter.Handler)
	http.Handle("/", prometheus.IndexHandler(metricsPath))

	_ = level.Info(Logger).Log("msg", "Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		_ = level.Error(Logger).Log("err", err)
		os.Exit(1)
	}
}
