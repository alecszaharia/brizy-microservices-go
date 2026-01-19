# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go monorepo** for Brizy microservices using **Go workspaces**. It follows a clean architecture pattern based
on the **go-kratos/kratos** framework for building cloud-native microservices.

### Workspace Structure

The repository uses Go 1.25 workspaces with three main modules:

- `contracts/` - Shared protobuf definitions and generated code (gRPC, Connect RPC, OpenAPI)
- `platform/` - Shared platform utilities (middleware, pagination, adapters)
- `services/{service-name}/` - {service-name} management microservice

### Architecture Pattern

Services follow **Clean Architecture** with dependency injection via Google Wire:

```
cmd/
  └── {service-name}/
      ├── main.go       # Entry point
      ├── wire.go       # Wire dependency definitions
      └── wire_gen.go   # Generated wire code
internal/
  ├── biz/            # Business logic layer (use cases)
  │   ├── interfaces.go   # Repository interfaces
  │   ├── models.go       # Business models
  │   └── {entity}.go     # Use case implementations
  ├── data/           # Data access layer (repositories)
  │   ├── data.go         # Database setup (GORM)
  │   ├── model/          # ORM entities
  │   └── repo/           # Repository implementations
  ├── service/        # Service layer (gRPC/HTTP handlers)
  │   ├── service.go      # Service struct
  │   ├── {entity}.go     # Handler implementations
  │   └── mapper.go       # DTO ↔ Business model conversions
  ├── server/         # Server setup (gRPC, HTTP)
  └── conf/           # Configuration protobuf definitions
```

**Layer dependencies**: `service` → `biz` → `data`

- Service layer depends on business layer
- Business layer defines interfaces, data layer implements them
- Wire handles dependency injection across all layers

## Development Commands

### Protobuf/API Contracts (Root Level)

```bash
# Generate code from proto files (gRPC, Connect RPC, OpenAPI)
make contracts-generate

# Lint protobuf files
make contracts-lint

# Check for breaking API changes against main branch
make contracts-breaking

# Format proto files
make contracts-format

# Full workflow: format, lint, and generate
make contracts-all
```

**Note**: Proto definitions are organized as:
- Service-specific: `api/service/{service}/v1/*.proto`

Generated code goes to `contracts/gen/` with the same directory structure.

### Service Development (services/{service-name}/)

```bash
cd services/{service-name}

# Generate Wire dependency injection code
make generate
# Note: Uses GOWORK=off to avoid workspace conflicts

# Build the service binary
make build
# Output: bin/symbol-service

# Run unit tests with race detection
make test

# Generate coverage report
make coverage
# Opens coverage.html in browser

# Generate config protobuf code
make config
# Note: buf generate for internal/conf/conf.proto
```

### Go Linting (golangci-lint)

The monorepo uses a hierarchical linting structure:

```bash
# Platform module
make lint-platform          # Lint platform code
make lint-platform-fix      # Lint with auto-fix

# Specific service
cd services/symbols
make lint                   # Lint service code
make lint-fix               # Lint with auto-fix

# All modules
make lint-all               # Lint platform + all services
make lint-all-fix           # Lint with auto-fix
```

**Configuration Files**:
- Root `.golangci.yml` - Platform module (scoped to platform/ via skip-dirs)
- `services/{service}/.golangci.yml` - Service-specific (copy of platform baseline)
- Contracts excluded (generated code)

📖 See [docs/golangci-lint.md](docs/golangci-lint.md) for detailed documentation.

### Running Tests

```bash
# Run tests for a specific service
cd services/{service-name} && go test ./internal/...

# Run single test
cd services/{service-name} && go test -run TestSymbolUseCase_Create ./internal/biz/

# Run tests with verbose output
cd services/{service-name} && go test -v ./internal/...
```

### Platform Module

The `platform/` module contains shared code:

- `events/` - Publisher/Subscriber interfaces for event-driven architecture
- `logger/` - Structured logging with Watermill integration
- `metrics/` - **Prometheus metrics export for observability** (HTTP, gRPC, pub/sub metrics)
- `middleware/` - Request ID middleware with context propagation
- `pagination/` - Offset-based pagination utilities
- `adapters/` - Common transformers and adapters

Platform code is imported by services as `platform/{package}` (e.g., `platform/events`, `platform/metrics`).

## Key Technical Details

### Wire Dependency Injection

Each layer exports a `ProviderSet`:

- `server.ProviderSet` - HTTP and gRPC server constructors
- `service.ProviderSet` - Service layer constructors
- `biz.ProviderSet` - Business logic constructors (use cases, validators)
- `data.ProviderSet` - Data access constructors (repos, DB connection)

After modifying `wire.go`, run `make generate` to regenerate `wire_gen.go`.

### Database (GORM)

Services use GORM for ORM. Data layer pattern:

- `internal/data/model/` - GORM entities with struct tags
- `internal/data/repo/` - Repository implementations satisfying `biz` interfaces
- `internal/data/common/transaction.go` - Transaction management utilities

### Configuration

Services use protobuf for configuration:

- `internal/conf/conf.proto` - Configuration schema
- `configs/config.yaml` - Runtime configuration
- Config loaded via Kratos config with env overrides (prefix: `KRATOS_`)

### Testing

Tests are co-located with implementation files (e.g., `symbols.go` → `symbols_test.go`).
Use `testify/mock` for mocking repository interfaces defined in `biz/interfaces.go`.

### Protobuf Tools

Required tools (install via `make init` in service directory):

- `buf` - Proto linting, breaking change detection, code generation
- `protoc-gen-go`, `protoc-gen-go-grpc` - Standard Go proto plugins
- `protoc-gen-go-http` - Kratos HTTP bindings
- `protoc-gen-openapi` - OpenAPI spec generation
- `wire` - Dependency injection code generation
- `protoc-gen-validate` - Proto validation rules

### API Design

APIs use:

- **gRPC** for service-to-service communication
- **HTTP/JSON** via Kratos HTTP bindings (mapped from gRPC via annotations)
- **Connect RPC** for browser-friendly gRPC
- **protoc-gen-validate** for request validation

Example from `symbols.proto`:

```protobuf
rpc CreateSymbol(CreateSymbolRequest) returns (CreateSymbolResponse) {
option (google.api.http) = {
post: "/v1/{service-name}"
    body: "*"
    };
    }
```

This generates both gRPC and HTTP handlers automatically.

## Event-Driven Architecture

Services implement event-driven patterns using Watermill for pub/sub messaging with RabbitMQ (AMQP).

### Worker Services

Each service can have a companion **worker** binary that processes asynchronous events:

```
services/{service-name}/
  ├── cmd/
  │   ├── {service-name}/        # Main HTTP/gRPC service
  │   └── {service-name}-worker/ # Event processing worker
  ├── internal/
  │   ├── handlers/              # Event handlers (business logic)
  │   ├── worker/                # Worker setup (Watermill router)
  │   └── ...
```

**Worker Architecture**:
- Workers run the Watermill message router
- Handlers delegate to business layer (biz) use cases
- Full context propagation (request_id, correlation_id) from publisher to subscriber
- Graceful shutdown with configurable timeout

### Platform Abstractions

Event interfaces are defined in `platform/events/`:

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, payload []byte) error
    Unwrap() message.Publisher
}

type Subscriber interface {
    Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
    Close() error
    Unwrap() message.Subscriber
}
```

These abstractions allow swapping message brokers (AMQP, Redis, Kafka) without changing business logic.

### Publisher/Subscriber Wrappers

Located in `services/{service-name}/internal/data/mq/`:

- **Publisher** (`publisher.go`) - Enriches messages with:
  - Context propagation (`msg.SetContext(ctx)`)
  - Correlation ID (generated or preserved)
  - Request ID (extracted from HTTP context)
  - Structured logging with full tracing

- **Subscriber** (`subscriber.go`) - Provides:
  - Lifecycle management
  - Structured logging
  - Access to underlying broker via `Unwrap()`

### Context Propagation

End-to-end tracing flows through the entire system:

```
HTTP Request → Middleware (request_id) → Use Case → Publisher
    ↓
Message (with context + metadata)
    ↓
Subscriber → Handler → Use Case (same request_id in logs)
```

**Best Practice**: Always use `logger.WithContext(ctx)` for distributed tracing.

### Documentation

For detailed information on extending the pub/sub system, adding new event types, and troubleshooting:

📖 See [docs/pubsub-architecture.md](docs/pubsub-architecture.md)

## Metrics and Observability

All services export Prometheus metrics for comprehensive observability.

### Metrics Infrastructure

**Location**: `platform/metrics/`

All services automatically export metrics on the `/metrics` endpoint (configurable) in Prometheus format.

**Configuration** (in `internal/conf/conf.proto`):

```protobuf
message Metrics {
  bool enabled = 1;              // Enable/disable metrics (default: true)
  string service_name = 2;       // Namespace for metrics (e.g., "symbols")
  string path = 3;               // Endpoint path (default: "/metrics")
  bool include_runtime = 4;      // Include Go runtime metrics (default: true)
}
```

**YAML Configuration** (in `configs/config.yaml`):

```yaml
metrics:
  enabled: true
  service_name: symbols
  path: /metrics
  include_runtime: true
```

### Available Metrics

**HTTP Server**:
- `{service}_http_requests_total{method, route, status}` - Request count
- `{service}_http_request_duration_seconds{method, route, status}` - Request latency

**gRPC Server**:
- `{service}_grpc_requests_total{service, method, status}` - Request count
- `{service}_grpc_request_duration_seconds{service, method, status}` - Request latency

**Watermill Pub/Sub**:
- `{service}_watermill_published_total{topic}` - Messages published
- `{service}_watermill_publish_duration_seconds{topic}` - Publish latency
- `{service}_watermill_publish_errors_total{topic}` - Publish errors
- `{service}_watermill_consumed_total{topic}` - Messages consumed
- `{service}_watermill_consume_duration_seconds{topic}` - Consume latency
- `{service}_watermill_consume_errors_total{topic}` - Consume errors
- `{service}_watermill_handler_acks_total{topic}` - Acknowledged messages
- `{service}_watermill_handler_nacks_total{topic}` - Nacked messages

**Runtime** (if enabled):
- `go_goroutines`, `go_memstats_*`, `go_gc_*`, `process_*`

**Build Info**:
- `{service}_build_info{version}` - Build version (always 1)

### Custom Business Metrics

Services can register custom metrics via the `metrics.Registry`:

```go
type MyUseCase struct {
    counter   prometheus.Counter
    histogram *prometheus.HistogramVec
}

func NewMyUseCase(registry *metrics.Registry) *MyUseCase {
    return &MyUseCase{
        counter: registry.NewCounter(
            "operations_total",
            "Total operations processed",
        ),
        histogram: registry.NewHistogramVec(
            "operation_duration_seconds",
            "Operation duration",
            []float64{0.01, 0.05, 0.1, 0.5, 1.0},
            []string{"operation_type"},
        ),
    }
}
```

**Naming Conventions**:
- All metrics automatically prefixed with service name
- Use `snake_case`
- Counters end with `_total`
- Durations end with `_seconds`
- Sizes end with `_bytes`
- Avoid unbounded label values (user IDs, request IDs)

### Integration Pattern

Services integrate metrics via Wire dependency injection:

1. Add `Metrics` config to `internal/conf/conf.proto`
2. Create `server.NewMetricsRegistry()` provider
3. Add `metrics.HTTPMiddleware()` and `metrics.GRPCMiddleware()` to middleware chain
4. Wrap pub/sub with `metrics.NewPublisherWithMetrics()` and `metrics.NewSubscriberWithMetrics()`
5. Register metrics handler at configured path
6. Regenerate Wire: `make generate`

**Reference Implementation**: `services/symbols/`

### Documentation

For detailed setup instructions, custom metrics, and troubleshooting:

📖 See [platform/metrics/README.md](platform/metrics/README.md)

### Mandatory actions:

- Always update the documentation in all files(CLAUDE.md, README.md, docs/ (if exists)). Add only big architectural
  changes that are important for developers to know