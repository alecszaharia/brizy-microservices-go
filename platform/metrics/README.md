# Platform Metrics

Prometheus metrics export for go-kratos microservices in the Brizy monorepo.

## Overview

The `platform/metrics` package provides standardized Prometheus metrics collection and export for all microservices. It automatically instruments HTTP/gRPC servers and Watermill pub/sub messaging with zero configuration, while supporting custom business metrics.

## Quick Start

### 1. Add Metrics Configuration to Your Service

In `internal/conf/conf.proto`:

```protobuf
message Metrics {
  bool enabled = 1;              // Enable/disable metrics export (default: true)
  string service_name = 2;       // Service namespace for metrics (e.g., "symbols")
  string path = 3;               // Endpoint path (default: "/metrics")
  bool include_runtime = 4;      // Include Go runtime metrics (default: true)
}

message Bootstrap {
    Server server = 1;
    Data data = 2;
    Metrics metrics = 3;  // Add metrics config
}
```

Run `make config` to regenerate protobuf code.

### 2. Configure in config.yaml

```yaml
metrics:
  enabled: true
  service_name: symbols
  path: /metrics
  include_runtime: true
```

### 3. Update Wire Providers

**In `internal/server/provider.go`:**

```go
import (
    "platform/metrics"
    "symbols/internal/conf/gen"
    "github.com/google/wire"
)

var ProviderSet = wire.NewSet(
    NewHTTPServer,
    NewGRPCServer,
    NewMetricsRegistry,
)

func NewMetricsRegistry(mc *conf.Metrics) *metrics.Registry {
    if mc == nil || !mc.Enabled {
        return nil
    }
    return metrics.NewRegistry(mc.ServiceName)
}
```

**In `internal/server/http.go`:**

```go
import (
    "platform/metrics"
    kratos_middleware "github.com/go-kratos/kratos/v2/middleware"
)

func NewHTTPServer(c *conf.Server, mc *conf.Metrics, reg *metrics.Registry, /* ... */) *http.Server {
    // Build middleware chain
    middlewares := []kratos_middleware.Middleware{
        recovery.Recovery(),
        ratelimit.Server(),
        middleware.RequestIDMiddleware(logger),
    }

    // Add metrics middleware if enabled
    if mc != nil && mc.Enabled && reg != nil {
        middlewares = append(middlewares, metrics.HTTPMiddleware(reg))
    }

    middlewares = append(middlewares,
        logging.Server(logger),
        validate.ProtoValidate(),
    )

    opts := []http.ServerOption{
        http.Middleware(middlewares...),
    }

    // ... other server options ...

    srv := http.NewServer(opts...)

    // Register metrics endpoint
    if mc != nil && mc.Enabled && reg != nil {
        metricsPath := mc.Path
        if metricsPath == "" {
            metricsPath = "/metrics"
        }
        srv.HandleFunc(metricsPath, metrics.NewMetricsHandler(reg))
    }

    // Register your services...
    return srv
}
```

**In `internal/server/grpc.go`:**

```go
func NewGRPCServer(c *conf.Server, mc *conf.Metrics, reg *metrics.Registry, /* ... */) *grpc.Server {
    middlewares := []kratos_middleware.Middleware{
        recovery.Recovery(),
        ratelimit.Server(),
        middleware.RequestIDMiddleware(logger),
    }

    // Add metrics middleware if enabled
    if mc != nil && mc.Enabled && reg != nil {
        middlewares = append(middlewares, metrics.GRPCMiddleware(reg))
    }

    // ... rest of middleware and options
}
```

**In `internal/data/data.go`:**

```go
import (
    "platform/events"
    "platform/metrics"
)

func NewEventPublisherWithMetrics(pub message.Publisher, mc *conf.Metrics, reg *metrics.Registry, logger log.Logger) events.Publisher {
    basePub := mq.NewEventPublisher(pub, logger)
    if mc != nil && mc.Enabled && reg != nil {
        return metrics.NewPublisherWithMetrics(basePub, reg)
    }
    return basePub
}

func NewEventSubscriberWithMetrics(sub message.Subscriber, mc *conf.Metrics, reg *metrics.Registry, logger log.Logger) events.Subscriber {
    baseSub := mq.NewEventSubscriber(sub, logger)
    if mc != nil && mc.Enabled && reg != nil {
        return metrics.NewSubscriberWithMetrics(baseSub, reg)
    }
    return baseSub
}
```

**Update `internal/data/provider.go` to use the new wrappers instead of direct `mq.NewEventPublisher/Subscriber`.**

### 4. Update Wire and Main

**In `cmd/{service}/wire.go`:**

```go
func wireApp(*conf.Server, *conf.Data, *conf.LogConfig, *conf.Metrics, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(/* ... your provider sets ... */))
}
```

**In `cmd/{service}/main.go`:**

```go
app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Log, bc.Metrics, logger)
```

### 5. Regenerate Wire

```bash
cd services/{service-name}
make generate
```

### 6. Access Metrics

Start your service and access metrics:

```bash
curl http://localhost:8000/metrics
```

## Available Metrics

### HTTP Server

- `{service}_http_requests_total{method, route, status}` - Total HTTP requests
- `{service}_http_request_duration_seconds{method, route, status}` - HTTP request latency histogram

**Important**: Uses route patterns (e.g., `/users/{id}`) NOT raw paths to prevent cardinality explosion.

### gRPC Server

- `{service}_grpc_requests_total{service, method, status}` - Total gRPC requests
- `{service}_grpc_request_duration_seconds{service, method, status}` - gRPC request latency histogram

### Watermill Pub/Sub

**Publisher:**
- `{service}_watermill_published_total{topic}` - Messages published successfully
- `{service}_watermill_publish_duration_seconds{topic}` - Publish latency histogram
- `{service}_watermill_publish_errors_total{topic}` - Publish errors

**Subscriber:**
- `{service}_watermill_consumed_total{topic}` - Messages consumed
- `{service}_watermill_consume_duration_seconds{topic}` - Consume/processing latency
- `{service}_watermill_consume_errors_total{topic}` - Consumption errors
- `{service}_watermill_handler_acks_total{topic}` - Acknowledged messages
- `{service}_watermill_handler_nacks_total{topic}` - Nacked (rejected) messages

### Runtime Metrics (if `include_runtime: true`)

- `go_goroutines` - Number of goroutines
- `go_memstats_*` - Memory statistics
- `go_gc_*` - Garbage collection stats
- `process_*` - Process metrics (CPU, file descriptors, etc.)

### Build Info

- `{service}_build_info{version}` - Build version (always set to 1)

## Custom Metrics

### Using the Registry

The metrics registry provides factory methods for creating custom metrics:

```go
type MyUseCase struct {
    requestCounter   prometheus.Counter
    queueGauge      prometheus.Gauge
    durationHist    *prometheus.HistogramVec
}

func NewMyUseCase(registry *metrics.Registry) *MyUseCase {
    return &MyUseCase{
        // Simple counter (no labels)
        requestCounter: registry.NewCounter(
            "processed_requests_total",
            "Total requests processed",
        ),

        // Simple gauge (no labels)
        queueGauge: registry.NewGauge(
            "queue_depth",
            "Current queue depth",
        ),

        // Histogram with labels
        durationHist: registry.NewHistogramVec(
            "operation_duration_seconds",
            "Operation duration in seconds",
            []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0}, // Custom buckets
            []string{"operation_type", "status"},
        ),
    }
}

func (uc *MyUseCase) ProcessRequest(ctx context.Context, opType string) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        status := "success"
        uc.durationHist.WithLabelValues(opType, status).Observe(duration)
    }()

    // Business logic...
    uc.requestCounter.Inc()
    uc.queueGauge.Set(42)

    return nil
}
```

### Counter with Labels

```go
statusCounter := registry.NewCounterVec(
    "operations_by_status_total",
    "Operations by status",
    []string{"status", "type"},
)

statusCounter.WithLabelValues("success", "create").Inc()
statusCounter.WithLabelValues("error", "update").Inc()
```

### Gauge with Labels

```go
connGauge := registry.NewGaugeVec(
    "active_connections",
    "Active connections by endpoint",
    []string{"endpoint"},
)

connGauge.WithLabelValues("database").Set(10)
connGauge.WithLabelValues("rabbitmq").Inc() // +1
connGauge.WithLabelValues("rabbitmq").Dec() // -1
```

## Naming Conventions

**All metrics are automatically prefixed with the service name** (e.g., `symbols_`).

Follow these conventions when creating custom metrics:

- Use `snake_case` for metric and label names
- Counters MUST end with `_total`
- Durations MUST end with `_seconds`
- Sizes MUST end with `_bytes`
- Avoid unbounded label values (user IDs, request IDs, timestamps)

**Examples:**
- ✅ `symbols_http_requests_total`
- ✅ `symbols_cache_hits_total`
- ✅ `symbols_job_duration_seconds`
- ✅ `symbols_message_size_bytes`
- ❌ `symbolsHTTPRequests` (not snake_case)
- ❌ `symbols_requests` (counter missing `_total`)
- ❌ `symbols_duration` (missing `_seconds` unit)

## Histogram Buckets

Default buckets are optimized for microservice latency:

```go
[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}
```

Most requests <100ms, p99 <1s. Customize buckets for your use case:

```go
histogram := registry.NewHistogramVec(
    "database_query_duration_seconds",
    "Database query duration",
    []float64{0.001, 0.01, 0.1, 1.0, 10.0}, // Longer queries
    []string{"query_type"},
)
```

## Prometheus Configuration

Example `prometheus.yml` for scraping services:

```yaml
scrape_configs:
  - job_name: 'symbols-service'
    static_configs:
      - targets: ['localhost:8000']
    metrics_path: /metrics
    scrape_interval: 15s

  - job_name: 'users-service'
    static_configs:
      - targets: ['localhost:8001']
    metrics_path: /metrics
    scrape_interval: 15s
```

## Performance

Metrics collection adds minimal overhead:

- **~50-100μs per request** (p99)
- Negligible memory impact with bounded cardinality
- Safe for production use

## Cardinality Management

**CRITICAL**: Avoid unbounded label values to prevent memory exhaustion.

### ✅ Good (Bounded)

```go
// Route patterns (limited set)
http_requests_total{route="/users/{id}"}

// Status codes (limited set)
http_requests_total{status="200"}

// Topics (known set)
watermill_published_total{topic="user.created"}

// Operations (known set)
cache_operations_total{operation="get"}
```

### ❌ Bad (Unbounded)

```go
// Raw paths create millions of unique metrics
http_requests_total{path="/users/12345"}
http_requests_total{path="/users/67890"}

// User IDs
requests_total{user_id="abc123"}

// Request IDs
requests_total{request_id="req-xyz"}

// Timestamps
events_total{timestamp="2025-01-11T10:00:00Z"}
```

**Rule of thumb**: Label values should have <1000 unique combinations per metric.

## Troubleshooting

### Metrics endpoint returns 404

- Check `metrics.enabled: true` in config
- Verify `metrics.path` matches your request path
- Ensure Wire is regenerated (`make generate`)

### No metrics appearing

- Check Wire is regenerated successfully
- Verify registry is passed to servers in Wire providers
- Check logs for initialization errors
- Ensure services are receiving traffic (metrics only appear after first request)

### High cardinality warnings from Prometheus

- Review your custom metrics for unbounded labels
- Use route patterns, not raw paths
- Limit label values to known sets
- Monitor unique time series count in Prometheus

### Worker metrics not appearing

For Phase 1 (MVP), workers don't expose their own metrics endpoints. Worker publish metrics are captured by the main service. Phase 2 will add standalone metrics servers for workers.

## Best Practices

### 1. Use Middleware for Request Metrics

Don't manually instrument HTTP/gRPC handlers. The middleware automatically captures all requests.

### 2. Keep Label Cardinality Low

```go
// ✅ Good: bounded label values
counter.WithLabelValues("success", "user_creation")

// ❌ Bad: unbounded user IDs
counter.WithLabelValues(userID) // DON'T DO THIS
```

### 3. Use Histograms for Latency

Histograms provide percentiles (p50, p95, p99) and are better than gauges for latency:

```go
histogram.Observe(duration.Seconds())
```

### 4. Increment Counters for Events

Counters should only increase (never decrease):

```go
// ✅ Good
successCounter.Inc()
failureCounter.Inc()

// ❌ Bad
totalCounter.Dec() // DON'T DO THIS - use gauge
```

### 5. Pre-create Metrics in Constructors

Create metrics once during initialization, not per-request:

```go
// ✅ Good: create in constructor
func NewService(reg *metrics.Registry) *Service {
    return &Service{
        counter: reg.NewCounter("requests_total", "Total requests"),
    }
}

// ❌ Bad: creates new metric per request
func (s *Service) Handle() {
    counter := s.reg.NewCounter(...) // DON'T DO THIS
}
```

## Future Enhancements (Phase 2)

Planned for future releases:

- Worker standalone metrics server on separate port
- gRPC stream interceptor support
- In-flight request gauges
- Build info from ldflags
- Custom exporters (beyond Prometheus)

## Testing

The metrics package has comprehensive test coverage (99.3%). Run tests:

```bash
cd platform
go test ./metrics/...
```

Check coverage:

```bash
go test -coverprofile=coverage.out ./metrics/...
go tool cover -html=coverage.out
```

## Architecture

The metrics package follows these patterns:

- **Decorator Pattern**: Wraps `events.Publisher/Subscriber` without modifying interfaces
- **Dependency Injection**: No global singletons, all dependencies injected via Wire
- **Registry Isolation**: Each service has its own registry to prevent conflicts
- **Middleware Composition**: Metrics collection via Kratos middleware chain

## Support

For questions or issues:

- See `CLAUDE.md` for monorepo conventions
- Check test files for usage examples
- Review integration in `services/symbols/` as reference implementation
