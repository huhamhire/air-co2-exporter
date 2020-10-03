package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Exporter struct {
	Registry  *prometheus.Registry
	GaugeTemp *prometheus.GaugeVec
	GaugeCo2  *prometheus.GaugeVec
	Handler   http.Handler
}

func NewExporter() *Exporter {
	e := Exporter{
		Registry: prometheus.NewRegistry(),
		GaugeTemp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "air_temp",
				Help: "Ambient Temperature (Tamb) in ℃.",
			},
			[]string{"tag"},
		),
		GaugeCo2: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "air_co2",
				Help: "Relative Concentration of CO2 (CntR) in ppm.",
			},
			[]string{"tag"},
		),
	}
	e.Registry.MustRegister(e.GaugeTemp)
	e.Registry.MustRegister(e.GaugeCo2)

	e.Handler = promhttp.HandlerFor(
		e.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
	return &e
}

const defaultTag = "default"

func (e *Exporter) SetTemp(value float64, tag string) {
	e.GaugeTemp.WithLabelValues(tag).Set(value)
}

func (e *Exporter) SetPpmCo2(value uint16, tag string) {
	e.GaugeCo2.WithLabelValues(tag).Set(float64(value))
}

func (e *Exporter) SetDefaultTemp(value float64) {
	e.SetTemp(value, defaultTag)
}

func (e *Exporter) SetDefaultPpmCo2(value uint16) {
	e.SetPpmCo2(value, defaultTag)
}

func IndexHandler(metricsPath *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Co2 Exporter</title></head>
			<body>
			<h1>Co2 Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
}
