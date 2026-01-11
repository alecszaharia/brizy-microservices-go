package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

var (
	// defaultHistogramBuckets are optimized for microservice latency.
	// Most requests <100ms, p99 <1s.
	defaultHistogramBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}
)

// HTTPMiddleware creates a Kratos middleware for recording HTTP request metrics.
func HTTPMiddleware(registry *Registry) middleware.Middleware {
	if registry == nil {
		// Return a no-op middleware if metrics are disabled
		return func(handler middleware.Handler) middleware.Handler {
			return handler
		}
	}

	// Create metrics
	requestsTotal := registry.NewCounterVec(
		"http_requests_total",
		"Total number of HTTP requests",
		[]string{"method", "route", "status"},
	)

	requestDuration := registry.NewHistogramVec(
		"http_request_duration_seconds",
		"HTTP request duration in seconds",
		defaultHistogramBuckets,
		[]string{"method", "route", "status"},
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

			// Extract route pattern and method
			route := tr.Operation()
			if route == "" {
				route = "unknown"
			}

			method := "unknown"
			// Check if this is HTTP transport
			if httpTr, ok := tr.(kratoshttp.Transporter); ok {
				httpReq := httpTr.Request()
				if httpReq != nil {
					method = httpReq.Method
				}
				// Use path template for better route pattern
				pathTemplate := httpTr.PathTemplate()
				if pathTemplate != "" {
					route = pathTemplate
				}
			}

			// Call the handler
			reply, err := handler(ctx, req)

			// Determine status code
			statusCode := 200
			if err != nil {
				// Default to 500 for errors, actual status may be set by error handling middleware
				statusCode = 500
			}
			status := strconv.Itoa(statusCode)

			// Record metrics
			duration := time.Since(startTime).Seconds()
			requestsTotal.WithLabelValues(method, route, status).Inc()
			requestDuration.WithLabelValues(method, route, status).Observe(duration)

			return reply, err
		}
	}
}
