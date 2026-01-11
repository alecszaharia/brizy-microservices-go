package server

import (
	"platform/build"
	"platform/metrics"
	conf "symbols/internal/conf/gen"
)

// NewMetricsRegistry creates a new metrics registry if metrics are enabled.
func NewMetricsRegistry(mc *conf.Metrics, buildInfo *build.ServiceBuildInfo) *metrics.Registry {
	if mc == nil || !mc.Enabled.Value {
		return nil
	}
	return metrics.NewRegistry(buildInfo)
}
