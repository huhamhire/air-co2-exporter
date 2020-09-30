package main

import (
	"co2-exporter/prometheus"
	"net/http"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
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
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("co2_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	//device := monitor.DefaultDevice()
	//m := monitor.NewDeviceMonitor(device)
	//_ = m.Connect()

	e := prometheus.NewExporter()

	e.SetDefaultTemp(60.2)

	_ = level.Info(logger).Log("msg", "Starting Co2 Exporter", "version", version.Info())

	http.Handle(*metricsPath, e.Handler)
	http.Handle("/", prometheus.IndexHandler(metricsPath))

	_ = level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
