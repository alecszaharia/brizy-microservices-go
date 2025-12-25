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

## Prerequisites

- **Go 1.25+**
- **Docker & Docker Compose**
- **buf** ([installation](https://buf.build/docs/installation/))
- **make**

Supported environments: macOS, Ubuntu

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
