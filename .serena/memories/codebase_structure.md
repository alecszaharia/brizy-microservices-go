# Codebase Structure

## Workspace Layout

The repository uses **Go 1.25 workspaces** with three main modules:

```
brizy-go-services/
├── contracts/           # Shared protobuf generated code
│   └── {service}/v1/    # Generated gRPC, Connect RPC, OpenAPI
├── platform/            # Shared platform utilities
│   ├── middleware/      # Request ID middleware
│   ├── pagination/      # Offset-based pagination
│   └── adapters/        # Common transformers
├── services/            # Microservices
│   └── {service}/         # Symbol management service
├── api/                 # Proto definitions
│   └── {service}/v1/    # Service API protobuf files
├── Makefile             # Root-level contracts commands
├── go.work              # Workspace definition
├── buf.yaml             # Buf configuration
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
│   │   └── repo/            # Repository implementations
│   ├── service/             # Service layer (gRPC/HTTP handlers)
│   │   ├── service.go       # Service struct
│   │   ├── {entity}.go      # Handler implementations
│   │   └── mapper.go        # DTO ↔ Business model conversions
│   ├── server/              # Server setup (gRPC, HTTP)
│   │   ├── grpc.go          # gRPC server configuration
│   │   └── http.go          # HTTP server configuration
│   └── conf/                # Configuration
│       └── conf.proto       # Config protobuf schema
├── configs/
│   └── config.yaml          # Runtime configuration
├── Makefile                 # Service-specific commands
├── go.mod                   # Service module definition
└── buf.gen.yaml             # Buf generation config
```

## Layer Dependencies

**Service → Biz → Data**

- Service layer depends on business layer (use cases)
- Business layer defines repository interfaces
- Data layer implements repository interfaces
- Wire handles dependency injection across all layers

## Wire ProviderSets

Each layer exports a ProviderSet for dependency injection:
- `server.ProviderSet` - HTTP and gRPC server constructors
- `service.ProviderSet` - Service layer constructors  
- `biz.ProviderSet` - Business logic constructors (use cases, validators)
- `data.ProviderSet` - Data access constructors (repos, DB connection)
