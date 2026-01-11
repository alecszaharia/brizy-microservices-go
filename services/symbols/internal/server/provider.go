package server

import (
	"platform/metrics"
	"symbols/internal/conf/gen"

	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	NewMetricsRegistry,
)

// NewMetricsRegistry creates a new metrics registry if metrics are enabled.
func NewMetricsRegistry(mc *conf.Metrics) *metrics.Registry {
	if mc == nil || !mc.Enabled {
		return nil
	}
	return metrics.NewRegistry(mc.ServiceName)
}
