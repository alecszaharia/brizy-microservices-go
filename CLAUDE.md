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

**Note**: Proto definitions are in `api/{service}/v1/*.proto`. Generated code goes to `contracts/{service}/v1/`.

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

- `middleware/` - Request ID middleware with context propagation
- `pagination/` - Offset-based pagination utilities
- `adapters/` - Common transformers and adapters

Platform code is imported by services as `brizy-go-platform/{package}`.

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

### Mandatory actions:

- Always update the documentation in all files(CLAUDE.md, README.md, docs/ (if exists)). Add only big architectural
  changes that are important for developers to know