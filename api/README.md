# API Definitions

This directory contains **Protocol Buffer (protobuf) definitions** for all microservices in the monorepo.

## Structure

```
api/
├── service/                            # Service-specific API definitions
│   └── {service-name}/                 # {service-name} service API definitions
│       └── v1/
│           ├── {service-name}.proto    # Service definitions (RPCs, messages)
│           └── error_reason.proto      # Service-specific error codes
└── common/                             # Shared protos across all services
    └── conf/                           # Shared configuration protos
        └── v1/
            └── conf.proto              # Common configuration definitions
```

### Directory Organization

- **`service/{service-name}/v1/`** - Service-specific API definitions versioned at v1
    - `{service-name}.proto` - gRPC service interfaces, request/response messages
    - `error_reason.proto` - Enum definitions for service-specific error codes - not mandatory
- **`common/conf/v1/`** - Shared configuration proto definitions used across all services

## Purpose

Proto files in this directory define:

- **gRPC service interfaces** - RPC methods and their request/response types
- **HTTP/JSON mappings** - REST endpoints via `google.api.http` annotations
- **Data contracts** - Shared message types between services
- **Validation rules** - Request validation via `protoc-gen-validate`

## Code Generation

Proto definitions are compiled into Go code in the `contracts/` module using `buf` and various protoc plugins.

### Available Commands (run from project root)

```bash
# Full workflow: format, lint, and generate
make contracts-all

# Generate gRPC, Connect RPC, and OpenAPI code
make contracts-generate

# Format proto files
make contracts-format

# Lint protobuf files
make contracts-lint

# Check for breaking API changes against main branch
make contracts-breaking
```

### Generated Output

Generated code is placed in `contracts/gen/` and includes:

- **`{service}.pb.go`** - Message type definitions (via `protoc-gen-go`)
- **`{service}_grpc.pb.go`** - gRPC service and client stubs (via `protoc-gen-go-grpc`)
- **`{service}_http.pb.go`** - HTTP bindings for Kratos (via `protoc-gen-go-http`)
- **`v1connect/{service}.connect.go`** - Connect RPC client/server code (via `protoc-gen-connectrpc`)
- **`api.json`** - OpenAPI v2 (Swagger) specification (via `protoc-gen-openapiv2`)

### Workflow

1. Define or modify proto files in `api/service/{service-name}/v1/` or `api/common/conf/v1/`
2. Run `make contracts-all` to format, lint, and generate code
3. Import generated code in service implementations:
   ```go
   import pb "contracts/gen/service/{service-name}/v1"
   import confpb "contracts/gen/common/conf/v1"
   ```

**WARNING**: After modifying proto definitions, you must re-run `make contracts-generate` (or `make contracts-all`) to
update generated code.

### CI Requirements

The CI pipeline enforces:

- Proto files must be formatted (`buf format`)
- Proto files must pass linting (`buf lint`)
- Changes must not break API contracts (`buf breaking`)

See [CLAUDE.md](/CLAUDE.md) for complete development workflow and architecture details.