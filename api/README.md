# API Definitions

This directory contains **Protocol Buffer (protobuf) definitions** for all microservices in the monorepo.

## Structure

```
api/
└── {service-name}/
    └── v1/
        ├── {service-name}.proto  # Service definitions (RPCs, messages)
        └── error_reason.proto    # Service-specific error codes
```

## Purpose

Proto files in this directory define:
- **gRPC service interfaces** - RPC methods and their request/response types
- **HTTP/JSON mappings** - REST endpoints via `google.api.http` annotations
- **Data contracts** - Shared message types between services
- **Validation rules** - Request validation via `protoc-gen-validate`

## Code Generation

Proto definitions are compiled into Go code in the `contracts/` module:

```bash
# Generate gRPC, Connect RPC, and OpenAPI code
make contracts-generate

# Lint and check for breaking changes
make contracts-lint
make contracts-breaking
```

Generated code is placed in `contracts/{service-name}/v1/`.

**WARNING**: after modifying proto definitions, you must re-run `make contracts-generate` to update generated code.


See [CLAUDE.md](/CLAUDE.md) for complete protobuf workflow.