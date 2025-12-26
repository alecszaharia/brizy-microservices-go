# Brizy Go Services

**Go monorepo** for Brizy microservices using Go 1.25 workspaces and the [go-kratos](https://go-kratos.dev/) framework.

## Architecture

```
brizy-go-services/
├── api/              # Protobuf definitions → See api/README.md
├── contracts/        # Generated Go code from protos → See contracts/README.md
├── platform/         # Shared utilities (middleware, pagination, adapters)
└── services/
    └── {service-name}/      # All service-specific code
```

Services follow **Clean Architecture** with layers: `service` → `biz` → `data`.

## Requirements

### Core Tools

- **Go 1.25+** - Primary programming language ([download](https://go.dev/dl/))
- **make** - Build automation (pre-installed on macOS/Linux)
- **git** - Version control (required for versioning and breaking change detection)
- **Docker & Docker Compose** - Container runtime ([installation](https://docs.docker.com/get-docker/))

### Protocol Buffers Tools

- **buf** - Protocol buffer management ([installation](https://buf.build/docs/installation/))

The following protoc plugins are required for service development and can be installed via `make init` in any service directory:

```bash
cd services/symbols
make init
```

This installs:
- `protoc-gen-go` - Go protobuf code generation
- `protoc-gen-go-grpc` - Go gRPC code generation
- `protoc-gen-go-http` - Kratos HTTP bindings
- `protoc-gen-openapi` - OpenAPI specification generation
- `protoc-gen-validate` - Protobuf validation rules
- `wire` - Dependency injection code generation
- `kratos` - Kratos framework CLI

**Note**: Remote buf plugins (Connect RPC, OpenAPI v2) are automatically managed by buf and don't require local installation.

### Supported Environments

macOS, Ubuntu

## Quick Start

### 1. Generate API Contracts

```bash
# Generate gRPC, Connect RPC, and OpenAPI code from protos
make contracts-generate

# Or run full workflow (format, lint, generate)
make contracts-all
```

### 2. Build & Run Services

```bash
# Build symbol service
cd services/symbols
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
make coverage   # Generate coverage report
```

## Documentation

- **[CLAUDE.md](CLAUDE.md)** - Complete development guide (architecture, commands, patterns)
- **[api/README.md](api/README.md)** - Protobuf definitions and validation
- **[contracts/README.md](contracts/README.md)** - Generated code structure and usage
- **Services** - Each service has its own README in `services/{service-name}/`

## Available Commands

Run `make help` to see all available targets.

## Workspace Modules

The Go workspace includes three modules:

- `contracts/` - Shared API contracts (auto-generated)
- `platform/` - Shared platform utilities
- `services/symbols/` - Symbol management service
