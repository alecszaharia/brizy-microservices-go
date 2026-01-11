<objective>
Implement comprehensive metrics export for the go-kratos monorepo, exposing both Kratos framework metrics (HTTP/gRPC server) and Watermill pub/sub metrics via a /metrics endpoint in Prometheus format.

This metrics infrastructure will be shared across all microservices in the monorepo, enabling standardized observability, performance monitoring, and operational insights for production deployments. The implementation must support custom business metrics and be configurable per service.
</objective>

<context>
This is a Go monorepo using go-kratos framework (v2) with Clean Architecture patterns. Services use Watermill for event-driven pub/sub messaging via RabbitMQ.

**Read project conventions:**
- @CLAUDE.md - Architecture overview, monorepo structure, platform module patterns
- @.claude/skills/path-finder/SKILL.md - File location conventions
- @platform/events/events.go - Pub/sub interfaces
- @platform/middleware/ - Example platform module structure

**Current structure:**
- `platform/` - Single Go module with shared utilities (events, logger, middleware, pagination, metrics)
- `services/{service}/` - Individual microservices following Clean Architecture
- Each service has: `cmd/{service}/` (main) and `cmd/{service}-worker/` (event worker)
- Services use Wire for dependency injection with ProviderSets
- Configuration via protobuf in `internal/conf/conf.proto`

**Existing patterns to follow:**
- Platform is a single module (`platform/go.mod`) with multiple packages
- Use structured logging with context propagation
- Configuration should be optional with sensible defaults
- Wire ProviderSets for dependency injection (see @services/symbols/cmd/symbols-worker/wire.go)
- No global singletons - use dependency injection for testability
</context>

<requirements>

## Platform Module Requirements (`platform/metrics/` package)

### 1. Core Metrics Infrastructure
Create `platform/metrics/` package within existing platform module:

**Prometheus Registry (Dependency Injected):**
- Registry instance per service (NOT global singleton)
- Support for custom metric registration (Counter, Gauge, Histogram)
- Service-level namespace isolation using service name
- Wire ProviderSet for dependency injection

**HTTP Handler:**
- Prometheus text exposition format handler for `/metrics` endpoint
- Include standard Go runtime metrics (goroutines, memory, GC)
- Include build info metric (simple version constant)
- ServerOption for Kratos HTTP server integration

**Metric Helpers:**
- Factory functions for creating metric vectors with standardized naming
- Helpers for recording HTTP request metrics (duration, status, route pattern)
- Helpers for recording gRPC request metrics (duration, status, method)
- Helpers for recording Watermill pub/sub metrics (published, consumed, errors)

### 2. Kratos Metrics Collection

**HTTP Server Metrics:**
- `http_requests_total{method, route, status}` - Counter of HTTP requests
  - **CRITICAL**: Use `route` (pattern like `/symbols/{id}`), NOT `path` (avoids cardinality explosion)
- `http_request_duration_seconds{method, route, status}` - Histogram of request duration

**gRPC Server Metrics:**
- `grpc_requests_total{service, method, status}` - Counter of gRPC requests
- `grpc_request_duration_seconds{service, method, status}` - Histogram of request duration

**Implementation approach:**
- Create Kratos middleware that intercepts HTTP/gRPC requests
- Extract route pattern from Kratos router (not raw path - prevents high cardinality)
- Record metrics before and after request handling
- Use prometheus.HistogramVec and observe duration
- Support gRPC unary interceptor (defer stream support to later phase)

**Histogram buckets:**
Use buckets optimized for microservice latency (most requests <100ms, p99 <1s):
- `[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0]` seconds

### 3. Watermill Metrics Collection

**Publisher Metrics:**
- `watermill_published_total{topic}` - Counter of messages published
- `watermill_publish_duration_seconds{topic}` - Histogram of publish duration
- `watermill_publish_errors_total{topic}` - Counter of publish errors

**Subscriber Metrics:**
- `watermill_consumed_total{topic}` - Counter of messages consumed
- `watermill_consume_duration_seconds{topic}` - Histogram of processing duration
- `watermill_consume_errors_total{topic}` - Counter of consumption errors
- `watermill_handler_acks_total{topic}` - Counter of acknowledged messages
- `watermill_handler_nacks_total{topic}` - Counter of nacked messages

**Implementation approach:**
- Wrap `platform/events.Publisher` and `platform/events.Subscriber` with metrics decorators
- Record metrics without modifying existing pub/sub interfaces (decorator pattern)
- Extract topic from Watermill messages
- Measure publish and consume latency
- Simple error counting (no error type classification)

### 4. Custom Metrics Support

**Metric Registry API:**
```go
// Registry allows services to create custom metrics
type Registry interface {
    // NewCounterVec creates a counter with labels
    NewCounterVec(name, help string, labelNames []string) *prometheus.CounterVec

    // NewGaugeVec creates a gauge with labels
    NewGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec

    // NewHistogramVec creates a histogram with custom buckets and labels
    NewHistogramVec(name, help string, buckets []float64, labelNames []string) *prometheus.HistogramVec

    // NewCounter creates a counter without labels (convenience method)
    NewCounter(name, help string) prometheus.Counter

    // NewGauge creates a gauge without labels (convenience method)
    NewGauge(name, help string) prometheus.Gauge
}
```

**Naming conventions to enforce:**
- All metrics prefixed with service namespace: `{service}_metric_name`
- Service namespace from config (e.g., "symbols", "users")
- Use snake_case for metric names
- Labels in snake_case
- Include `_total` suffix for counters
- Include `_seconds` suffix for duration histograms
- Include `_bytes` suffix for size metrics

**Note:** Prometheus SDK will validate label formats. No custom validation needed.

### 5. Configuration Support

**Service configuration proto:**

In each service's `internal/conf/conf.proto`, add:
```protobuf
message Metrics {
  bool enabled = 1;              // Enable/disable metrics export (default: true)
  string service_name = 2;       // Service namespace for metrics (e.g., "symbols")
  string path = 3;               // Endpoint path (default: "/metrics")
  bool include_runtime = 4;      // Include Go runtime metrics (default: true)
}
```

**Why duplicate instead of shared proto:**
- Platform is a single module, services import packages (not proto)
- Proto sharing adds complexity (import paths, generation)
- 4-field message is simple enough to duplicate
- Each service can customize if needed

**Service configuration YAML:**
```yaml
metrics:
  enabled: true
  service_name: symbols  # Used as metric prefix
  path: /metrics
  include_runtime: true
```

### 6. Worker Service Support

**For Phase 1 (MVP):**
- Workers share the same metrics approach as main services
- If worker has no HTTP server, metrics won't be exposed (acceptable for MVP)
- Worker metrics (publish-only) are captured by main service anyway

**For Phase 2 (future enhancement):**
- Add standalone metrics server on separate port for workers
- Configuration: `metrics.addr: :9090`

</requirements>

<implementation>

## Step-by-Step Implementation

### Phase 1: Platform Metrics Package

**Add to existing `platform/` module (no new go.mod):**

1. **Update `platform/go.mod`**
   - Add Prometheus dependency if not present:
   ```
   require (
       github.com/prometheus/client_golang v1.18.0
       // ... existing dependencies
   )
   ```

2. **File: `platform/metrics/registry.go`**
   - Define Registry interface and implementation
   - Constructor: `NewRegistry(serviceName string) *prometheus.Registry`
   - Implement metric factory methods (NewCounterVec, NewGaugeVec, NewHistogramVec)
   - Implement convenience methods (NewCounter, NewGauge)
   - Enforce naming conventions: prepend service name to metrics
   - Register Go runtime collectors: `prometheus.NewGoCollector()`, `prometheus.NewProcessCollector()`
   - Register build info metric with simple version constant

3. **File: `platform/metrics/handler.go`**
   - HTTP handler function using `promhttp.HandlerFor(registry, opts)`
   - Function signature: `NewMetricsHandler(registry *prometheus.Registry) http.Handler`
   - Configure compression, timeout, error handling

4. **File: `platform/metrics/server.go`**
   - Kratos ServerOption for auto-registering metrics handler
   - `MetricsServerOption(registry *prometheus.Registry, path string) http.ServerOption`
   - Registers metrics handler at specified path (default: "/metrics")

5. **File: `platform/metrics/http.go`**
   - Kratos HTTP middleware for request metrics
   - Function: `HTTPMiddleware(registry *prometheus.Registry, serviceName string) middleware.Middleware`
   - Extract route pattern from Kratos operation context (NOT raw URL path)
   - Record: duration histogram, request counter
   - Handle errors gracefully (don't fail requests if metrics fail)

6. **File: `platform/metrics/grpc.go`**
   - Kratos gRPC unary interceptor: `UnaryServerInterceptor(registry *prometheus.Registry, serviceName string)`
   - Extract service and method from gRPC context
   - Record: duration histogram, request counter
   - **Note:** Stream interceptor deferred to Phase 2

7. **File: `platform/metrics/watermill.go`**
   - Publisher decorator: `NewPublisherWithMetrics(pub events.Publisher, registry *prometheus.Registry, serviceName string) events.Publisher`
   - Subscriber decorator: `NewSubscriberWithMetrics(sub events.Subscriber, registry *prometheus.Registry, serviceName string) events.Subscriber`
   - Extract topic from messages
   - Measure latency, count successes/errors
   - Simple error counting (no error type classification)

8. **File: `platform/metrics/config.go`**
   - Go struct for metrics configuration:
   ```go
   type Config struct {
       Enabled        bool
       ServiceName    string
       Path           string
       IncludeRuntime bool
   }

   func DefaultConfig(serviceName string) *Config {
       return &Config{
           Enabled:        true,
           ServiceName:    serviceName,
           Path:           "/metrics",
           IncludeRuntime: true,
       }
   }
   ```

9. **File: `platform/metrics/provider.go`**
   ```go
   package metrics

   import "github.com/google/wire"

   // ProviderSet is metrics providers for Wire
   var ProviderSet = wire.NewSet(
       NewRegistry,
       NewMetricsHandler,
       HTTPMiddleware,
       UnaryServerInterceptor,
   )
   ```

10. **File: `platform/metrics/version.go`**
    - Simple version constant (no ldflags complexity):
    ```go
    package metrics

    const (
        Version = "1.0.0"  // Update manually or via CI
    )

    func registerBuildInfo(registry *prometheus.Registry, serviceName string) {
        buildInfo := prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: serviceName + "_build_info",
                Help: "Build information",
            },
            []string{"version"},
        )
        buildInfo.WithLabelValues(Version).Set(1)
        registry.MustRegister(buildInfo)
    }
    ```

### Phase 2: Service Integration (Main Service)

**Update `services/symbols/`:**

1. **File: `services/symbols/internal/conf/conf.proto`**
   - Add Metrics message (duplicate, not shared):
   ```protobuf
   message Metrics {
     bool enabled = 1;
     string service_name = 2;
     string path = 3;
     bool include_runtime = 4;
   }

   message Bootstrap {
       Server server = 1;
       Data data = 2;
       Metrics metrics = 3;  // Add metrics config
   }
   ```
   - Run: `cd services/symbols && make config` to regenerate

2. **File: `services/symbols/internal/server/http.go`**
   - Import `platform/metrics`
   - Modify `NewHTTPServer` signature to accept metrics config and registry
   - Add metrics middleware to middleware stack
   - Add metrics route if enabled:
   ```go
   func NewHTTPServer(c *conf.Server, mc *conf.Metrics, reg *prometheus.Registry, symbolService *service.SymbolService, logger log.Logger) *http.Server {
       var opts = []http.ServerOption{
           http.Middleware(
               recovery.Recovery(),
               ratelimit.Server(),
               middleware.RequestIDMiddleware(logger),
               metrics.HTTPMiddleware(reg, mc.ServiceName), // Add metrics middleware
               logging.Server(logger),
               validate.ProtoValidate(),
           ),
       }

       // ... existing CORS, network, addr, timeout setup ...

       srv := http.NewServer(opts...)

       // Register metrics handler if enabled
       if mc.Enabled {
           srv.Route(mc.Path).GET(metrics.NewMetricsHandler(reg))
       }

       v1.RegisterSymbolsServiceHTTPServer(srv, symbolService)
       return srv
   }
   ```

3. **File: `services/symbols/internal/server/grpc.go`**
   - Import `platform/metrics`
   - Add gRPC interceptor if metrics enabled:
   ```go
   func NewGRPCServer(c *conf.Server, mc *conf.Metrics, reg *prometheus.Registry, symbolService *service.SymbolService, logger log.Logger) *grpc.Server {
       var opts = []grpc.ServerOption{
           grpc.Middleware(
               recovery.Recovery(),
               // ... other middleware ...
           ),
       }

       if mc.Enabled {
           opts = append(opts,
               grpc.UnaryInterceptor(metrics.UnaryServerInterceptor(reg, mc.ServiceName)),
           )
       }

       // ... rest of setup ...
   }
   ```

4. **File: `services/symbols/internal/data/data.go`**
   - Update ProviderSet to include metrics-wrapped pub/sub:
   ```go
   var ProviderSet = wire.NewSet(
       NewData,
       repo.NewSymbolRepo,
       NewEventPublisher,  // Wraps with metrics
       NewEventSubscriber, // Wraps with metrics
   )

   func NewEventPublisher(pub message.Publisher, reg *prometheus.Registry, mc *conf.Metrics, logger log.Logger) events.Publisher {
       basePub := mq.NewEventPublisher(pub, logger)
       if mc != nil && mc.Enabled && reg != nil {
           return metrics.NewPublisherWithMetrics(basePub, reg, mc.ServiceName)
       }
       return basePub
   }

   func NewEventSubscriber(sub message.Subscriber, reg *prometheus.Registry, mc *conf.Metrics, logger log.Logger) events.Subscriber {
       baseSub := mq.NewEventSubscriber(sub, logger)
       if mc != nil && mc.Enabled && reg != nil {
           return metrics.NewSubscriberWithMetrics(baseSub, reg, mc.ServiceName)
       }
       return baseSub
   }
   ```

5. **File: `services/symbols/internal/server/provider.go`**
   - Create if doesn't exist
   - Add ProviderSet including metrics registry:
   ```go
   var ProviderSet = wire.NewSet(
       NewHTTPServer,
       NewGRPCServer,
       NewMetricsRegistry,
   )

   func NewMetricsRegistry(mc *conf.Metrics) *prometheus.Registry {
       if mc == nil || !mc.Enabled {
           return nil
       }
       return metrics.NewRegistry(mc.ServiceName)
   }
   ```

6. **File: `services/symbols/cmd/symbols/wire.go`**
   - Add server.ProviderSet if not already present
   - Wire will inject metrics config and registry into servers

7. **File: `services/symbols/configs/config.yaml`**
   ```yaml
   server:
     http:
       addr: :8000
       timeout: 1s
     grpc:
       addr: :9000
       timeout: 1s

   metrics:
     enabled: true
     service_name: symbols
     path: /metrics
     include_runtime: true

   data:
     # ... existing data config ...
   ```

8. **Regenerate Wire:**
   ```bash
   cd services/symbols
   make generate
   ```

### Phase 3: Testing

**Create comprehensive tests:**

1. **File: `platform/metrics/registry_test.go`**
   - Test metric creation with proper namespacing
   - Test duplicate metric registration errors
   - Test namespace prepending
   - Test registry isolation (multiple registries don't conflict)

2. **File: `platform/metrics/http_test.go`**
   - Test middleware records metrics correctly
   - Test route pattern extraction (not raw path)
   - Mock Kratos operation context
   - Test duration histogram buckets
   - Test error cases (middleware doesn't crash requests)

3. **File: `platform/metrics/grpc_test.go`**
   - Test unary interceptor records metrics
   - Mock gRPC context with service/method
   - Test status code labeling

4. **File: `platform/metrics/watermill_test.go`**
   - Test publisher decorator records publish metrics
   - Test subscriber decorator records consume metrics
   - Test topic extraction from messages
   - Test latency measurement

5. **File: `platform/metrics/handler_test.go`**
   - Test metrics handler returns Prometheus text format
   - Test content-type header: `text/plain; version=0.0.4; charset=utf-8`
   - Test runtime metrics presence
   - Test build info metric

6. **File: `platform/metrics/config_test.go`**
   - Test default config values
   - Test config struct

**Integration tests** (in service directory):

7. **File: `services/symbols/internal/server/http_test.go`**
   - Test `/metrics` endpoint returns valid Prometheus format
   - Test metrics endpoint when disabled
   - Test HTTP request metrics increment
   - Parse Prometheus output, verify metric names

### Phase 4: Documentation

1. **File: `platform/metrics/README.md`**
   ```markdown
   # Platform Metrics

   Prometheus metrics export for go-kratos microservices.

   ## Quick Start

   ### 1. Add metrics config to your service

   In `internal/conf/conf.proto`:
   ```protobuf
   message Metrics {
     bool enabled = 1;
     string service_name = 2;
     string path = 3;
     bool include_runtime = 4;
   }

   message Bootstrap {
       // ... existing fields
       Metrics metrics = 3;
   }
   ```

   ### 2. Configure in config.yaml

   ```yaml
   metrics:
     enabled: true
     service_name: myservice
     path: /metrics
     include_runtime: true
   ```

   ### 3. Update Wire providers

   Add metrics registry and wrap pub/sub (see implementation section).

   ### 4. Access metrics

   ```bash
   curl http://localhost:8000/metrics
   ```

   ## Available Metrics

   ### HTTP Server
   - `{service}_http_requests_total{method, route, status}` - Request count
   - `{service}_http_request_duration_seconds{method, route, status}` - Request latency

   ### gRPC Server
   - `{service}_grpc_requests_total{service, method, status}` - Request count
   - `{service}_grpc_request_duration_seconds{service, method, status}` - Request latency

   ### Watermill Pub/Sub
   - `{service}_watermill_published_total{topic}` - Messages published
   - `{service}_watermill_publish_duration_seconds{topic}` - Publish latency
   - `{service}_watermill_publish_errors_total{topic}` - Publish errors
   - `{service}_watermill_consumed_total{topic}` - Messages consumed
   - `{service}_watermill_consume_duration_seconds{topic}` - Consume latency
   - `{service}_watermill_consume_errors_total{topic}` - Consume errors
   - `{service}_watermill_handler_acks_total{topic}` - Acknowledged messages
   - `{service}_watermill_handler_nacks_total{topic}` - Nacked messages

   ### Runtime (if enabled)
   - `go_goroutines` - Number of goroutines
   - `go_memstats_*` - Memory statistics
   - `go_gc_*` - Garbage collection stats

   ### Build Info
   - `{service}_build_info{version}` - Build version (always 1)

   ## Custom Metrics

   ```go
   import "platform/metrics"

   type MyUseCase struct {
       counter   prometheus.Counter
       histogram *prometheus.HistogramVec
   }

   func NewMyUseCase(registry *prometheus.Registry) *MyUseCase {
       // Simple counter
       counter := registry.NewCounter(
           "operations_total",
           "Total operations processed",
       )

       // Counter with labels
       statusCounter := registry.NewCounterVec(
           "operations_by_status_total",
           "Operations by status",
           []string{"status"},
       )

       // Gauge
       queueDepth := registry.NewGauge(
           "queue_depth",
           "Current queue depth",
       )

       // Histogram with custom buckets
       histogram := registry.NewHistogramVec(
           "operation_duration_seconds",
           "Operation duration",
           []float64{0.01, 0.05, 0.1, 0.5, 1.0},
           []string{"operation_type"},
       )

       return &MyUseCase{
           counter:   counter,
           histogram: histogram,
       }
   }

   func (uc *MyUseCase) ProcessOperation(ctx context.Context, opType string) {
       start := time.Now()
       defer func() {
           duration := time.Since(start).Seconds()
           uc.histogram.WithLabelValues(opType).Observe(duration)
       }()

       // ... business logic ...
       uc.counter.Inc()
   }
   ```

   ## Naming Conventions

   - All metrics prefixed with service name: `{service}_metric_name`
   - Use `snake_case` for metric and label names
   - Counters end with `_total`
   - Durations end with `_seconds`
   - Sizes end with `_bytes`

   ## Performance

   Metrics collection adds ~50-100μs overhead per request (p99).

   ## Prometheus Configuration

   Example `prometheus.yml`:
   ```yaml
   scrape_configs:
     - job_name: 'symbols-service'
       static_configs:
         - targets: ['localhost:8000']
       metrics_path: /metrics
       scrape_interval: 15s
   ```

   ## Troubleshooting

   **Metrics endpoint returns 404:**
   - Check `metrics.enabled: true` in config
   - Check `metrics.path` matches your request path

   **No metrics appearing:**
   - Check Wire is regenerated (`make generate`)
   - Check registry is passed to servers in Wire

   **High cardinality warning:**
   - Avoid using unbounded labels (user IDs, request IDs)
   - Use route patterns, not raw paths
   - Limit label values to known set

   ## Future Enhancements (Phase 2)

   - Worker metrics server (separate port)
   - gRPC stream interceptor
   - In-flight request gauges
   - Build info from ldflags
   ```

2. **Update `CLAUDE.md`**
   - Add section on metrics infrastructure:
   ```markdown
   ### Platform Utilities

   **Location**: `platform/{package}/`

   **Packages**:
   - `platform/events` - Publisher/Subscriber interfaces for event-driven architecture
   - `platform/logger` - Structured logging with Watermill integration
   - `platform/metrics` - **NEW**: Prometheus metrics export for observability
   - `platform/middleware` - Request ID middleware with context propagation
   - `platform/pagination` - Offset-based pagination utilities

   **Import Examples**:
   ```go
   import "platform/events"
   import "platform/logger"
   import "platform/metrics"
   import "platform/middleware"
   import "platform/pagination"
   ```

   ## Metrics Infrastructure

   All services export Prometheus metrics on `/metrics` endpoint (configurable).

   **Configuration** (in `configs/config.yaml`):
   ```yaml
   metrics:
     enabled: true
     service_name: symbols  # Used as metric prefix
     path: /metrics
     include_runtime: true
   ```

   **Available metrics**: HTTP, gRPC, Watermill pub/sub, Go runtime

   See `platform/metrics/README.md` for detailed documentation.
   ```

</implementation>

<constraints>

**Why these constraints matter:**

1. **Cardinality Protection**: Prometheus stores all unique label combinations in memory. Unbounded labels (raw URL paths, user IDs, request IDs) create millions of time series, causing memory exhaustion. This is the #1 cause of Prometheus outages.

2. **Dependency Injection over Globals**: Global singletons make testing impossible (parallel tests conflict), hide dependencies, and create initialization order bugs. DI makes dependencies explicit and testable.

3. **Route Patterns not Paths**: HTTP paths like `/users/123`, `/users/456` create unbounded cardinality. Route patterns like `/users/{id}` are bounded and safe.

4. **Single Platform Module**: Platform is one module (`platform/go.mod`) with multiple packages. This simplifies dependency management and avoids module versioning issues.

5. **Performance**: Metrics add overhead to every request. Target <100μs p99 overhead.

**Specific constraints:**

- **DO NOT** use raw URL paths as labels (use route patterns)
- **DO NOT** create global singleton registry (use DI)
- **DO NOT** add user IDs, request IDs, session IDs as labels (unbounded cardinality)
- **DO NOT** use milliseconds for durations (Prometheus convention is seconds)
- **DO NOT** create submodules in platform (single go.mod)
- **DO** enforce metric naming conventions (service prefix, suffixes)
- **DO** use dependency injection for registry
- **DO** include comprehensive tests (>80% coverage target)
- **DO** document all metrics in README
- **DO** use standard histogram buckets: `[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0]`

</constraints>

<output>
Create the following files with exact paths:

**Platform metrics package (within platform module):**
- Update `./platform/go.mod` (add Prometheus dependency)
- `./platform/metrics/registry.go`
- `./platform/metrics/handler.go`
- `./platform/metrics/server.go`
- `./platform/metrics/http.go`
- `./platform/metrics/grpc.go`
- `./platform/metrics/watermill.go`
- `./platform/metrics/config.go`
- `./platform/metrics/provider.go`
- `./platform/metrics/version.go`
- `./platform/metrics/README.md`

**Tests:**
- `./platform/metrics/registry_test.go`
- `./platform/metrics/http_test.go`
- `./platform/metrics/grpc_test.go`
- `./platform/metrics/watermill_test.go`
- `./platform/metrics/handler_test.go`
- `./platform/metrics/config_test.go`

**Service integration (symbols as reference):**
- Modify `./services/symbols/internal/conf/conf.proto` (add Metrics message)
- Generate protobuf: run `cd services/symbols && make config`
- Modify `./services/symbols/internal/server/http.go` (add metrics middleware + handler)
- Modify `./services/symbols/internal/server/grpc.go` (add metrics interceptor)
- Create `./services/symbols/internal/server/provider.go` (Wire ProviderSet with metrics)
- Modify `./services/symbols/internal/data/data.go` (wrap pub/sub with metrics)
- Modify `./services/symbols/cmd/symbols/wire.go` (add server.ProviderSet)
- Modify `./services/symbols/configs/config.yaml` (add metrics config)
- Regenerate Wire: run `cd services/symbols && make generate`

**Documentation:**
- Update `./CLAUDE.md` (add platform/metrics section)
</output>

<research>
Before implementation, examine these files to understand existing patterns:

1. **Platform structure:**
   - @platform/go.mod - Check existing dependencies
   - @platform/events/events.go - Publisher/Subscriber interfaces for decorator pattern
   - @platform/middleware/ - Middleware patterns
   - @platform/logger/ - ProviderSet pattern

2. **Service structure:**
   - @services/symbols/internal/server/http.go - HTTP server setup and middleware stack
   - @services/symbols/internal/server/grpc.go - gRPC server setup
   - @services/symbols/internal/data/mq/publisher.go - Current publisher wrapper pattern
   - @services/symbols/internal/data/mq/subscriber.go - Current subscriber wrapper pattern
   - @services/symbols/internal/conf/conf.proto - Config proto structure

3. **Wire integration:**
   - @services/symbols/cmd/symbols-worker/wire.go - Wire setup example
   - Look for ProviderSet patterns in existing code

4. **Kratos specifics:**
   - How to extract route pattern from HTTP requests (not raw path)
   - Kratos middleware signature and usage
   - gRPC interceptor patterns in Kratos
</research>

<verification>
Before declaring complete, verify your implementation:

**Functionality:**
1. Start symbols service: `cd services/symbols && go run ./cmd/symbols`
2. Verify `/metrics` endpoint responds: `curl http://localhost:8000/metrics`
3. Check Prometheus format: response starts with `# HELP` and `# TYPE`
4. Make HTTP request: `curl http://localhost:8000/v1/symbols`
5. Verify `symbols_http_requests_total{method="GET",route="/v1/symbols",status="200"}` increments
6. Check route label uses pattern, not raw path
7. Publish event and verify `symbols_watermill_published_total` increments
8. Check runtime metrics: `go_goroutines`, `go_memstats_alloc_bytes`
9. Check build info: `symbols_build_info{version="1.0.0"} 1`

**Testing:**
1. Run platform tests: `cd platform && go test -v ./metrics/...`
2. All tests pass
3. Check coverage: `go test -coverprofile=coverage.out ./metrics/... && go tool cover -func=coverage.out`
4. Verify coverage >80%

**Integration:**
1. Verify service starts with metrics disabled: set `metrics.enabled: false` in config
2. Verify service starts with metrics enabled
3. Check no errors in startup logs
4. Verify Wire compilation: `cd services/symbols && make generate`
5. Verify protobuf generation: `cd services/symbols && make config`

**Performance:**
1. Benchmark HTTP middleware: `go test -bench=BenchmarkHTTPMiddleware -benchmem`
2. Verify overhead <100μs per request (p99)
3. Check histogram buckets cover actual latency distribution

**Documentation:**
1. README has installation steps
2. README lists all available metrics
3. README has custom metrics code example that compiles
4. CLAUDE.md updated with metrics section

**Production readiness:**
1. Metrics endpoint returns correct content-type header
2. No sensitive data (passwords, tokens) in metric labels
3. Cardinality is bounded (no unbounded labels)
4. Route patterns used, not raw paths
</verification>

<success_criteria>
1. ✅ `/metrics` endpoint returns valid Prometheus format with correct content-type
2. ✅ HTTP metrics use route patterns (not raw paths) to avoid cardinality explosion
3. ✅ gRPC metrics include unary interceptor
4. ✅ Watermill pub/sub metrics collected via decorator pattern
5. ✅ Registry uses dependency injection (not global singleton)
6. ✅ Services can register custom metrics via clean API
7. ✅ Metrics configurable via protobuf config (duplicated per service)
8. ✅ All tests pass with >80% coverage target
9. ✅ Naming conventions enforced (service prefix, suffixes)
10. ✅ Documentation comprehensive with runnable examples
11. ✅ symbols service integration works end-to-end
12. ✅ No breaking changes to existing interfaces
13. ✅ Performance overhead <100μs p99
14. ✅ Build info metric with simple version constant
15. ✅ Wire ProviderSet included for easy integration
16. ✅ Platform remains single module (no submodules)
17. ✅ Protobuf generation commands documented and work
18. ✅ No in-flight metrics (simplified)
19. ✅ No Summary metric type (Counter, Gauge, Histogram only)
20. ✅ No custom label validation (Prometheus SDK handles it)
</success_criteria>

<best_practices>
**Prometheus best practices:**
- Use seconds for durations (not milliseconds)
- Use `_total` suffix for counters
- Use `_seconds` suffix for duration histograms
- Keep label cardinality low (<1000 unique combinations per metric)
- Use histogram buckets matching your SLOs
- Include units in metric names (e.g., `_bytes`, `_seconds`)
- Use route patterns, not raw paths
- Avoid high-cardinality labels (user IDs, request IDs, session IDs, timestamps)

**Go-Kratos patterns:**
- Use middleware for cross-cutting concerns
- Follow Wire dependency injection (no globals)
- Extract route patterns from operation context
- Use structured logging with context
- Follow Clean Architecture layer separation

**Watermill patterns:**
- Use decorator pattern to avoid modifying interfaces
- Extract topic from messages for labeling
- Simple error counting (no complex error classification)

**Performance optimization:**
- Pre-allocate metric collectors during initialization
- Use prometheus.Observer pattern for histograms (no allocations)
- Avoid string concatenation in hot path
- Cache metric collectors, don't create on each request
</best_practices>
