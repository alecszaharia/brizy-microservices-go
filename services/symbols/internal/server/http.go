package server

import (
	v1 "contracts/gen/symbols/v1"
	"platform/middleware"
	"symbols/internal/conf/gen"
	"symbols/internal/service"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

func NewHTTPServer(c *conf.Server, symbolService *service.SymbolService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			ratelimit.Server(),
			middleware.RequestIDMiddleware(logger),
			logging.Server(logger),
			validate.ProtoValidate(),
		),
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
