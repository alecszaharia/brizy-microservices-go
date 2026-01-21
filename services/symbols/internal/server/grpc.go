// Package server configures HTTP and gRPC servers with middleware.
package server

import (
	v1 "contracts/gen/service/symbols/v1"
	"platform/metrics"
	"platform/middleware"
	"symbols/internal/conf/gen"
	"symbols/internal/service"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	kratos_middleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, mc *conf.Metrics, reg *metrics.Registry, symbolService *service.SymbolService, logger log.Logger) *grpc.Server {
	// Build middleware chain
	middlewares := []kratos_middleware.Middleware{
		recovery.Recovery(),
		ratelimit.Server(),
		middleware.RequestIDMiddleware(logger),
	}

	// Add metrics middleware if enabled
	if mc != nil && mc.Enabled.Value && reg != nil {
		middlewares = append(middlewares, metrics.GRPCMiddleware(reg))
	}

	middlewares = append(middlewares,
		logging.Server(logger),
		validate.ProtoValidate(),
	)

	var opts = []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterSymbolsServiceServer(srv, symbolService)
	return srv
}
