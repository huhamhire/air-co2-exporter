package prometheus

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsExposition(t *testing.T) {
	for _, format := range []struct {
		name   string
		accept string
	}{
		{"Prometheus", "text/plain;version=0.0.4"},
		{"OpenMetrics", "application/openmetrics-text;version=1.0.0"},
	} {
		t.Run(format.name, func(t *testing.T) {
			e := NewExporter()
			e.SetLabelTag("office")
			e.SetTemp(23.5)
			e.SetPpmCo2(800)
			e.SetRh(45.25)
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			req.Header.Set("Accept", format.accept)
			res := httptest.NewRecorder()
			e.Handler.ServeHTTP(res, req)
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body)
			}
			co2 := `air_co2{tag="office"} 800`
			if format.name == "OpenMetrics" {
				co2 += ".0"
			}
			for _, line := range []string{
				`air_temp{tag="office"} 23.5`,
				co2,
				`air_rh{tag="office"} 45.25`,
				"# TYPE air_temp gauge",
				"# TYPE air_co2 gauge",
				"# TYPE air_rh gauge",
			} {
				if !strings.Contains(res.Body.String(), line+"\n") {
					t.Errorf("missing metric line %q in %s", line, res.Body)
				}
			}
			if format.name == "OpenMetrics" && !strings.HasSuffix(res.Body.String(), "# EOF\n") {
				t.Error("missing OpenMetrics EOF marker")
			}
		})
	}
}

func TestDefaultLabelAndIndex(t *testing.T) {
	e := NewExporter()
	e.SetPpmCo2(600)
	res := httptest.NewRecorder()
	e.Handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(res.Body.String(), `air_co2{tag="default"} 600`) {
		t.Fatalf("default label missing: %s", res.Body)
	}
	path := "/custom-metrics"
	res = httptest.NewRecorder()
	IndexHandler(&path).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(res.Body.String(), `href="/custom-metrics"`) {
		t.Fatalf("custom metrics link missing: %s", res.Body)
	}
}
