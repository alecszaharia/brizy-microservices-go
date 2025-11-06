package middleware

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// CacheControl returns a middleware that adds cache control headers
func CacheControl(maxAge int) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			context.WithValue(ctx, "Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
			if rw, ok := http.ResponseWriterFromServerContext(ctx); ok {
				rw.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
			}
			return handler(ctx, req)
		}
	}
}
