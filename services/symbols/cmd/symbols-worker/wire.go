//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

//go:generate go run github.com/google/wire/cmd/wire

import (
	platform_logger "platform/logger"
	"platform/metrics"
	"symbols/internal/biz"
	"symbols/internal/conf/gen"
	"symbols/internal/data"
	"symbols/internal/handlers"
	"symbols/internal/worker"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// NewWorkerMetricsRegistry creates a metrics registry for the worker (returns nil for MVP).
func NewWorkerMetricsRegistry(mc *conf.Metrics) *metrics.Registry {
	// For Phase 1 (MVP), workers don't expose metrics endpoints
	// Workers' publish metrics are captured by the main service anyway
	return nil
}

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.LogConfig, *conf.Metrics, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		platform_logger.ProviderSet,
		worker.ProviderSet,
		handlers.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		NewWorkerMetricsRegistry,
		newApp,
	))
}
