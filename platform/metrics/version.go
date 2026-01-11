package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// Version is the current version of the metrics package.
	// Update manually or via CI.
	Version = "1.0.0"
)

// registerBuildInfo registers a build_info metric with version information.
func registerBuildInfo(registry *prometheus.Registry, serviceName string) {
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: serviceName + "_build_info",
			Help: "Build information",
		},
		[]string{"version"},
	)
	buildInfo.WithLabelValues(Version).Set(1)
	registry.MustRegister(buildInfo)
}
