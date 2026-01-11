package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
)

const RequestIdHeader = "X-Request-ID"
const RequestIdKey = "request_id"

type requestIDKey struct{}

func RequestIDMiddleware(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {

			var requestID string
			tr, hasTransport := transport.FromServerContext(ctx)

			if hasTransport {
				requestID = tr.RequestHeader().Get(RequestIdHeader)
			}

			l := log.NewHelper(logger).WithContext(ctx)
			if requestID == "" || !isValidRequestID(requestID, l) {
				requestID = generateId(l)
			}

			// Add to context
			ctx = context.WithValue(ctx, requestIDKey{}, requestID)

			// Add to response header if transport available
			if hasTransport {
				tr.ReplyHeader().Set(RequestIdHeader, requestID)
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

func generateId(logger *log.Helper) string {
	requestID := uuid.New().String()
	logger.Debugf("Generated new request ID: %s", requestID)
	return requestID
}

func isValidRequestID(id string, logger *log.Helper) bool {
	// validate id
	if err := uuid.Validate(id); err != nil {
		logger.Warnf("request id is invalid: %v", err)
		return false
	}

	return true
}
