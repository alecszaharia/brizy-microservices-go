package metrics

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/status"
)

// GRPCMiddleware creates a Kratos middleware for recording gRPC request metrics.
// This works as both a middleware and can be adapted for gRPC interceptor.
func GRPCMiddleware(registry *Registry) middleware.Middleware {
	if registry == nil {
		// Return a no-op middleware if metrics are disabled
		return func(handler middleware.Handler) middleware.Handler {
			return handler
		}
	}

	// Create metrics
	requestsTotal := registry.NewCounterVec(
		"grpc_requests_total",
		"Total number of gRPC requests",
		[]string{"service", "method", "status"},
	)

	requestDuration := registry.NewHistogramVec(
		"grpc_request_duration_seconds",
		"gRPC request duration in seconds",
		defaultHistogramBuckets,
		[]string{"service", "method", "status"},
	)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			startTime := time.Now()

			// Get transport info
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				// No transport info, just call handler
				return handler(ctx, req)
			}

			// Extract service and method from operation
			// Operation format: /package.Service/Method
			operation := tr.Operation()
			service := "unknown"
			method := "unknown"

			// Parse operation to get service and method
			service, method = parseGRPCOperation(operation)

			// Call the handler
			reply, err := handler(ctx, req)

			// Extract gRPC status code
			statusCode := "OK"
			if err != nil {
				st, _ := status.FromError(err)
				statusCode = st.Code().String()
			}

			// Record metrics
			duration := time.Since(startTime).Seconds()
			requestsTotal.WithLabelValues(service, method, statusCode).Inc()
			requestDuration.WithLabelValues(service, method, statusCode).Observe(duration)

			return reply, err
		}
	}
}

// parseGRPCOperation parses a gRPC operation string to extract service and method.
// Operation format: /package.Service/Method or just Service/Method
func parseGRPCOperation(operation string) (service, method string) {
	// Default values
	service = "unknown"
	method = "unknown"

	if operation == "" {
		return
	}

	// Remove leading slash if present
	if operation[0] == '/' {
		operation = operation[1:]
	}

	// Split by last slash to separate service and method
	lastSlash := -1
	for i := len(operation) - 1; i >= 0; i-- {
		if operation[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash > 0 {
		service = operation[:lastSlash]
		method = operation[lastSlash+1:]
	} else {
		// No slash found, treat entire string as method
		method = operation
	}

	return
}
