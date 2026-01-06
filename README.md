# Brizy Go Services

**Go monorepo** for Brizy microservices using Go 1.25 workspaces and the [go-kratos](https://go-kratos.dev/) framework.

## Requirements

### Core Tools

- **Go 1.25+** - Primary programming language
- **make** - Build automation (pre-installed on macOS/Linux)
- **git** - Version control (required for versioning and breaking change detection)
- **Docker & Docker Compose** - Container runtime ([installation](https://docs.docker.com/get-docker/))

### Protocol Buffers Tools

- **buf** - Protocol buffer management ([installation](https://buf.build/docs/installation/))

The following protoc plugins are required for service development and can be installed via `make init`:

```bash
make init
```

This installs:

- `protoc-gen-go` - Go protobuf code generation
- `protoc-gen-go-grpc` - Go gRPC code generation
- `protoc-gen-validate` - Protobuf validation rules
- `protoc-gen-go-http` - Kratos HTTP bindings
- `protoc-gen-openapi` - OpenAPI specification generation
- `protoc-gen-connect-go` - Connect RPC code generation
- `protoc-gen-openapiv2` - OpenAPI v2 (Swagger) specification generation
- `wire` - Dependency injection code generation
- `kratos` - Kratos framework CLI

## Quick Start

## Workspace Layout

The repository uses **Go 1.25 workspaces** with three main modules:

```
brizy-go-services/
├── api/                 # Proto definitions
│   ├── service/         # Service-specific API definitions
│   │   └── symbols/v1/  # Symbols service protobuf files
│   │       ├── symbols.proto        # Service definitions (RPCs, messages)
│   │       └── error_reason.proto   # Service-specific error codes
│   └── common/          # Shared protos across all services
│       └── conf/v1/     # Shared configuration protos
│           └── conf.proto           # Common configuration definitions
├── contracts/           # Shared protobuf generated code
│   ├── go.mod                    # Contracts module definition
│   ├── gen/                      # Generated code directory
│   │   ├── service/             # Service-specific generated code
│   │   │   └── symbols/v1/      # Generated gRPC, Connect RPC, HTTP for symbols
│   │   │       ├── symbols.pb.go           # Message definitions
│   │   │       ├── symbols_grpc.pb.go      # gRPC service/client stubs
│   │   │       ├── symbols_http.pb.go      # Kratos HTTP bindings
│   │   │       └── v1connect/              # Connect RPC generated code
│   │   │           └── symbols.connect.go  # Connect RPC service/client
│   │   ├── common/              # Common/shared generated code
│   │   │   └── conf/v1/         # Generated config protos
│   │   │       └── conf.pb.go           # Configuration structs
│   │   └── api.json             # OpenAPI v2 (Swagger) specification
│   └── README.md                 # Contracts documentation
├── platform/            # Shared platform utilities
│   ├── go.mod                    # Platform module definition
│   ├── logger/          # Centralized logger factory with trace/request ID support
│   ├── middleware/      # Request ID middleware with context propagation
│   ├── pagination/      # Offset-based pagination utilities
│   └── event/           # Event handling utilities
├── services/            # Microservices
│   ├── symbols/         # Symbol management service
│   └── test/            # Test service (if applicable)
├── .github/             # GitHub workflows and configurations
├── Makefile             # Root-level contracts commands
├── go.work              # Workspace definition
├── go.work.sum          # Workspace checksums
├── buf.yaml             # Buf configuration (root)
├── buf.gen.yaml         # Buf generation config
├── buf.lock             # Buf dependency lock file
├── docker-compose.yml   # Service orchestration
└── CLAUDE.md            # Development guide for AI assistants
```

## Service Architecture (Clean Architecture Pattern)

Each service follows this structure:

```
services/{service-name}/
├── cmd/
│   └── {service-name}/
│       ├── main.go          # Entry point
│       ├── wire.go          # Wire dependency definitions
│       └── wire_gen.go      # Generated wire code (auto-generated)
├── internal/
│   ├── biz/                 # Business logic layer (use cases)
│   │   ├── interfaces.go    # Repository interfaces
│   │   ├── models.go        # Business models
│   │   ├── validator.go     # Business validation
│   │   ├── errors.go        # Business errors
│   │   ├── {entity}.go      # Use case implementations
│   │   ├── {entity}_test.go # Unit tests
│   │   ├── events/          # Event publishing (if applicable)
│   │   │   └── publisher.go # Event publisher implementation
│   │   ├── biz.go           # Wire ProviderSet
│   │   └── README.md        # Business layer documentation
│   ├── data/                # Data access layer (repositories)
│   │   ├── data.go          # Database setup (GORM)
│   │   ├── model/           # GORM entities
│   │   │   └── {entity}.go  # ORM model definitions
│   │   ├── repo/            # Repository implementations
│   │   │   ├── {entity}.go      # Repository implementing biz interface
│   │   │   └── {entity}_test.go # Repository tests
│   │   ├── mq/              # Message queue integration (if applicable)
│   │   │   ├── publisher.go     # MQ publisher implementation
│   │   │   └── consumer.go      # MQ consumer implementation
│   │   ├── common/          # Shared data utilities
│   │   │   └── transaction.go   # Transaction management
│   │   └── README.md        # Data layer documentation
│   ├── service/             # Service layer (gRPC/HTTP handlers)
│   │   ├── service.go       # Service struct
│   │   ├── {entity}.go      # Handler implementations
│   │   ├── {entity}_test.go # Service handler tests
│   │   ├── mapper.go        # DTO ↔ Business model conversions
│   │   ├── mapper_test.go   # Mapper tests
│   │   ├── events/          # Event handling (if applicable)
│   │   └── README.md        # Service layer documentation
│   ├── server/              # Server setup (gRPC, HTTP)
│   │   ├── grpc.go          # gRPC server configuration
│   │   ├── http.go          # HTTP server configuration
│   │   ├── event.go         # Event server setup (if applicable)
│   │   └── server.go        # Wire ProviderSet
│   └── conf/                # Configuration
│       ├── conf.proto       # Config protobuf schema
│       └── conf.pb.go       # Generated config code
├── configs/
│   └── config.yaml          # Runtime configuration
├── bin/                     # Compiled binaries (auto-generated)
│   └── {service-name}       # Service executable
├── Makefile                 # Service-specific commands
├── Dockerfile               # Production container image
├── Dockerfile.debug         # Debug container image (optional)
├── go.mod                   # Service module definition
├── go.sum                   # Go module checksums
├── buf.yaml                 # Service-specific buf config (optional)
└── buf.gen.yaml             # Service-specific buf generation (optional)
```

Services follow **Clean Architecture** with layers: `service` → `biz` → `data`.


### Generate API Contracts only when necessary (Only there was a change in the proto files)

Always commit the generated code in contracts. Failing to commit the generated code will cause the CI to fail.

```bash
# Install protoc plugins and other dependencies
make init

# Generate gRPC, Connect RPC, and OpenAPI code from protos
make contracts-generate

# Or run full workflow (format, lint, generate)
make contracts-all
```

### Build & Run Services

```bash
# Build symbol service
cd services/{service-name}}
make build

# Run with Docker Compose
docker-compose up
```

## Development Workflow

### Proto Changes

Commands for working with proto files:

```bash
make contracts-format     # Format proto files
make contracts-lint       # Lint protos
make contracts-breaking   # Check for breaking changes vs main
make contracts-generate   # Generate Go code
```

### Service Development

Command for working with a specific service:

```bash
cd services/{service-name}

make generate   # Generate Wire dependency injection code
make build      # Build binary → bin/service-name
make test       # Run tests with race detection
make coverage   # Generate coverage report
```

## Documentation

- **[CLAUDE.md](CLAUDE.md)** - Complete development guide (architecture, commands, patterns)
- **[docs/debugging-goland.md](docs/debugging-goland.md)** - GoLand IDE debugging guide (local & remote)
- **[api/README.md](api/README.md)** - Protobuf definitions and validation
- **[contracts/README.md](contracts/README.md)** - Generated code structure and usage
- **Services** - Each service has its own README in `services/{service-name}/`

## Available Commands

Run `make help` to see all available targets.

## Workspace Modules

The Go workspace includes modules:

- `contracts/` - Shared API contracts (auto-generated)
- `platform/` - Shared platform utilities
- `services/{service-name}/` - {service-name} management service
