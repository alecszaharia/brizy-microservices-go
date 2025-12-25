# Contracts Module

This module contains **generated Go code** from the protobuf definitions in the [`api/`](/api) directory. It serves as the shared contracts layer for all microservices.

## Structure

```
contracts/
└── {service-name}/
    └── v1/
        ├── *.pb.go              # Protobuf message types
        ├── *_grpc.pb.go         # gRPC service interfaces
        ├── *_http.pb.go         # Kratos HTTP bindings
        └── v1connect/
            └── *.connect.go     # Connect RPC code
```

## Purpose

This module provides:
- **Type-safe API contracts** - Shared message types and service interfaces
- **gRPC clients and servers** - Auto-generated from proto service definitions
- **HTTP handlers** - REST endpoints mapped from gRPC methods
- **Connect RPC support** - Browser-friendly gRPC alternative

## Usage

Services import this module to access API contracts:

```go
import (
    symbolsv1 "brizy-go-contracts/symbols/v1"
    "brizy-go-contracts/symbols/v1/v1connect"
)
```

## Generation

Code is generated from proto files in [`api/`](/api):

```bash
# From repository root
make contracts-generate
```

**WARNING**: All files in this directory are auto-generated. Do not edit manually - changes will be overwritten. Modify the proto definitions in [`api/`](/api) instead.

See [api/README.md](/api/README.md) for proto definition details.