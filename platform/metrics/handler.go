package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsHandler creates an HTTP handler for Prometheus metrics exposition.
// Returns a standard http.Handler that serves metrics in Prometheus text format.
func NewMetricsHandler(registry *Registry) http.Handler {
	if registry == nil {
		// Return a handler that serves empty metrics
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Metrics disabled\n"))
		})
	}

	return promhttp.HandlerFor(
		registry.Unwrap(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}
