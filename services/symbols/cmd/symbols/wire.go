//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

//go:generate go run github.com/google/wire/cmd/wire

import (
	platform_build_info "platform/build"
	platform_logger "platform/logger"
	"symbols/internal/biz"
	"symbols/internal/conf/gen"
	"symbols/internal/data"
	"symbols/internal/server"
	"symbols/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*platform_build_info.ServiceBuildInfo, *conf.Server, *conf.Data, *conf.LogConfig, *conf.Metrics, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(platform_logger.ProviderSet, server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
