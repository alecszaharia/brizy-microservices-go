package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
)

const RequestHeader = "X-Request-ID"

type requestIDKey struct{}

func RequestIDMiddleware() middleware.Middleware {

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {

			var requestID string

			if tr, ok := transport.FromServerContext(ctx); ok {
				requestID = tr.RequestHeader().Get(RequestHeader)
			}

			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add to context
			ctx = context.WithValue(ctx, requestIDKey{}, requestID)

			// Add to Kratos log context
			// Add to response header if transport available
			if tr, ok := transport.FromServerContext(ctx); ok {
				tr.ReplyHeader().Set(RequestHeader, requestID)
			}

			return handler(ctx, req)
		}
	}
}

// RequestID returns a log.Valuer that extracts the request ID from context.
// This is used with log.With() to automatically include request ID in all logs.
func RequestID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
			return rid
		}
		return ""
	}
}
