# Codebase Structure

## Workspace Layout

The repository uses **Go 1.25 workspaces** with three main modules:

```
brizy-go-services/
├── api/                 # Proto definitions
│   ├── symbols/v1/      # Symbols service protobuf files
│   │   ├── symbols.proto      # Service definitions (RPCs, messages)
│   │   └── error_reason.proto # Service-specific error codes
│   └── conf/v1/         # Shared configuration protos
│       └── servers.proto      # Common server configurations
├── contracts/           # Shared protobuf generated code
│   └── symbols/v1/      # Generated gRPC, Connect RPC, OpenAPI
│       ├── symbols.pb.go          # Message definitions
│       ├── symbols_grpc.pb.go     # gRPC service/client stubs
│       ├── symbols_http.pb.go     # Kratos HTTP bindings
│       ├── symbols.pb.validate.go # Validation code
│       └── symbols.openapi.yaml   # OpenAPI specification
├── platform/            # Shared platform utilities
│   ├── middleware/      # Request ID middleware with context propagation
│   ├── pagination/      # Offset-based pagination utilities
│   └── adapters/        # Common transformers and adapters
├── services/            # Microservices
│   └── symbols/         # Symbol management service
├── Makefile             # Root-level contracts commands
├── go.work              # Workspace definition
├── buf.yaml             # Buf configuration (root)
├── buf.gen.yaml         # Buf generation config
└── docker-compose.yml   # Service orchestration
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
│   │   └── biz.go           # Wire ProviderSet
│   ├── data/                # Data access layer (repositories)
│   │   ├── data.go          # Database setup (GORM)
│   │   ├── model/           # GORM entities
│   │   │   └── {entity}.go  # ORM model definitions
│   │   └── repo/            # Repository implementations
│   │       └── {entity}.go  # Repository implementing biz interface
│   ├── service/             # Service layer (gRPC/HTTP handlers)
│   │   ├── service.go       # Service struct
│   │   ├── {entity}.go      # Handler implementations
│   │   └── mapper.go        # DTO ↔ Business model conversions
│   ├── server/              # Server setup (gRPC, HTTP)
│   │   ├── grpc.go          # gRPC server configuration
│   │   ├── http.go          # HTTP server configuration
│   │   └── server.go        # Wire ProviderSet
│   └── conf/                # Configuration
│       └── conf.proto       # Config protobuf schema
├── configs/
│   └── config.yaml          # Runtime configuration
├── Makefile                 # Service-specific commands
├── go.mod                   # Service module definition
└── buf.gen.yaml             # Service-specific buf config (optional)
```

## Layer Dependencies

**Service → Biz → Data**

- The service layer depends on the business layer (use cases)
- Business layer defines repository interfaces
- Data layer implements repository interfaces
- Wire handles dependency injection across all layers

**Dependency Rule**: Dependencies point inward (outer → inner)
- NO reverse dependencies allowed
- NO layer skipping (service cannot directly access data)

## Wire ProviderSets

Each layer exports a ProviderSet for dependency injection:

- `server.ProviderSet` - HTTP and gRPC server constructors
- `service.ProviderSet` - Service layer constructors  
- `biz.ProviderSet` - Business logic constructors (use cases, validators)
- `data.ProviderSet` - Data access constructors (repos, DB connection)

All ProviderSets are wired together in `cmd/{service-name}/wire.go`.

## Key Directories

- `api/` - Source of truth for API contracts (proto files)
- `contracts/` - Generated code (DO NOT manually edit)
- `platform/` - Shared utilities imported as `brizy-go-platform/{package}`
- `services/` - Individual microservices with Clean Architecture
- `internal/` - Service-private code (cannot be imported by other services)
