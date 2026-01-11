package server

import (
	v1 "contracts/gen/symbols/v1"
	"platform/metrics"
	"platform/middleware"
	"symbols/internal/conf/gen"
	"symbols/internal/service"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

func NewHTTPServer(c *conf.Server, mc *conf.Metrics, reg *metrics.Registry, symbolService *service.SymbolService, logger log.Logger) *http.Server {
	// Build middleware chain
	middlewares := []kratosmiddleware.Middleware{
		recovery.Recovery(),
		ratelimit.Server(),
		middleware.RequestIDMiddleware(logger),
	}

	// Add metrics middleware if enabled
	if mc != nil && mc.Enabled.Value && reg != nil {
		middlewares = append(middlewares, metrics.HTTPMiddleware(reg))
	}

	middlewares = append(middlewares,
		logging.Server(logger),
		validate.ProtoValidate(),
	)

	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
	}

	// Configure CORS if specified in config
	if c.Http.Cors != nil {
		corsOpts := buildCORSOptions(c.Http.Cors)
		opts = append(opts, http.Filter(handlers.CORS(corsOpts...)))
	}

	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)

	// Register metrics endpoint if enabled
	if mc != nil && mc.Enabled.Value && reg != nil {
		metricsPath := mc.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		srv.Handle(metricsPath, metrics.NewMetricsHandler(reg))
	}

	v1.RegisterSymbolsServiceHTTPServer(srv, symbolService)
	return srv
}

// NewHTTPServer new an HTTP server.
// buildCORSOptions constructs CORS options from configuration.
func buildCORSOptions(corsConfig *conf.CORS) []handlers.CORSOption {
	var opts []handlers.CORSOption

	if len(corsConfig.AllowedOrigins) > 0 {
		opts = append(opts, handlers.AllowedOrigins(corsConfig.AllowedOrigins))
	}

	if len(corsConfig.AllowedMethods) > 0 {
		opts = append(opts, handlers.AllowedMethods(corsConfig.AllowedMethods))
	}

	if len(corsConfig.AllowedHeaders) > 0 {
		opts = append(opts, handlers.AllowedHeaders(corsConfig.AllowedHeaders))
	}

	if len(corsConfig.ExposedHeaders) > 0 {
		opts = append(opts, handlers.ExposedHeaders(corsConfig.ExposedHeaders))
	}

	if corsConfig.AllowCredentials.GetValue() {
		opts = append(opts, handlers.AllowCredentials())
	}

	if corsConfig.MaxAge != nil {
		maxAgeSeconds := int(corsConfig.MaxAge.AsDuration().Seconds())
		opts = append(opts, handlers.MaxAge(maxAgeSeconds))
	}

	return opts
}
