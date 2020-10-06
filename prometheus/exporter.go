package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Exporter object
type Exporter struct {
	Registry  *prometheus.Registry
	GaugeTemp *prometheus.GaugeVec
	GaugeCo2  *prometheus.GaugeVec
	GaugeRh   *prometheus.GaugeVec
	Handler   http.Handler
	labelTag  string
}

// NewExporter - create a new Exporter instance
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
		GaugeRh: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "air_rh",
				Help: "Relative Humidity(RH).",
			},
			[]string{"tag"},
		),
		labelTag: "default",
	}
	e.Registry.MustRegister(e.GaugeTemp)
	e.Registry.MustRegister(e.GaugeCo2)
	e.Registry.MustRegister(e.GaugeRh)

	e.Handler = promhttp.HandlerFor(
		e.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
	return &e
}

// SetLabelTag - set tag label for the exporter instance
func (e *Exporter) SetLabelTag(tag string) {
	e.labelTag = tag
}

func (e *Exporter) setTemp(value float64, tag string) {
	e.GaugeTemp.WithLabelValues(tag).Set(value)
}

func (e *Exporter) setPpmCo2(value uint16, tag string) {
	e.GaugeCo2.WithLabelValues(tag).Set(float64(value))
}

func (e *Exporter) setRh(value float64, tag string) {
	e.GaugeRh.WithLabelValues(tag).Set(value)
}

// SetTemp - set temperature value
func (e *Exporter) SetTemp(value float64) {
	e.setTemp(value, e.labelTag)
}

// SetPpmCo2 - set CO2 value
func (e *Exporter) SetPpmCo2(value uint16) {
	e.setPpmCo2(value, e.labelTag)
}

// SetRh - set RH value
func (e *Exporter) SetRh(value float64) {
	e.setRh(value, e.labelTag)
}

// IndexHandler - default http handler for index
func IndexHandler(metricsPath *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Air Co2 Exporter</title></head>
			<body>
			<h1>Air Co2 Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
}
