---
name: path-finder
description: Provides correct file paths for brizy-go-services monorepo based on Clean Architecture conventions and workspace structure. Use when you need to reference or locate files in the codebase.
---

<objective>
Provide instant, accurate file paths for all main components in the brizy-go-services Go monorepo, following Clean Architecture patterns and go-kratos framework conventions.
</objective>

<workspace_structure>
The project uses **Go 1.25 workspaces** with three main modules:

**Root Level:**
```
/Users/alex/GolandProjects/brizy-go-services/
├── go.work                    # Workspace definition
├── Makefile                   # Contracts generation commands
├── CLAUDE.md                  # Project instructions
├── README.md                  # Project overview
├── docker-compose.yml         # Service orchestration
├── buf.yaml                   # Buf configuration (root)
├── buf.gen.yaml              # Buf generation config (root)
├── api/                       # Proto definitions
├── contracts/                 # Generated proto code
├── platform/                  # Shared utilities
└── services/                  # Microservices
```

**Workspace Modules:**
1. `contracts/` - Generated protobuf code
2. `platform/` - Shared platform utilities
3. `services/symbols/` - Symbol service
</workspace_structure>

<api_and_contracts>
**Proto Definitions (source):**
```
api/{service}/v1/{service}.proto
api/symbols/v1/symbols.proto
```

**Generated Contracts:**
```
contracts/{service}/v1/
contracts/symbols/v1/
  ├── {service}_grpc.pb.go        # gRPC service
  ├── {service}.pb.go             # Protocol buffers
  ├── {service}_http.pb.go        # HTTP bindings
  └── {service}.swagger.json      # OpenAPI spec
```

**Commands:**
- Generate: `make contracts-generate` (from root)
- Lint: `make contracts-lint` (from root)
- Format: `make contracts-format` (from root)
</api_and_contracts>

<platform_module>
**Shared Platform Code:**
```
platform/
├── go.mod                     # Platform module definition
├── middleware/
│   └── requestid/
│       ├── requestid.go       # Request ID middleware
│       └── context.go         # Context helpers
├── pagination/
│   ├── pagination.go          # Pagination utilities
│   └── meta.go               # Pagination metadata
└── adapters/
    └── ...                    # Common transformers
```

**Import Path:** `brizy-go-platform/{package}`
</platform_module>

<service_structure>
**Symbol Service (Clean Architecture):**

**Entry Point & Wire:**
```
services/symbols/cmd/symbols/
├── main.go                    # Service entry point
├── wire.go                    # Wire dependency definitions
└── wire_gen.go               # Generated wire code (auto-generated)
```

**Configuration:**
```
services/symbols/internal/conf/
├── conf.proto                 # Config schema (protobuf)
└── conf.pb.go                # Generated config code

services/symbols/configs/
└── config.yaml               # Runtime configuration
```

**Business Logic Layer (Biz):**
```
services/symbols/internal/biz/
├── biz.go                    # Wire ProviderSet
├── interfaces.go             # Repository interfaces (SymbolRepo, SymbolUseCase)
├── models.go                 # Business models (Symbol, ListSymbolsOptions)
├── validator.go              # Business validation
├── errors.go                 # Business errors
├── symbols.go                # Symbol use case implementation
└── symbols_test.go           # Use case unit tests
```

**Data Access Layer (Data):**
```
services/symbols/internal/data/
├── data.go                   # Database setup, Wire ProviderSet
├── model/
│   └── symbol.go             # GORM entities
├── repo/
│   └── symbol.go             # Repository implementations
└── common/
    └── transaction.go        # Transaction utilities
```

**Service Layer (Handlers):**
```
services/symbols/internal/service/
├── service.go                # Service struct, Wire ProviderSet
├── symbols.go                # gRPC/HTTP handlers
└── mapper.go                 # DTO ↔ Business model mapping
```

**Server Setup:**
```
services/symbols/internal/server/
├── server.go                 # Wire ProviderSet
├── grpc.go                   # gRPC server configuration
└── http.go                   # HTTP server configuration
```

**Build Output:**
```
services/symbols/bin/
└── symbols                   # Built binary
```

**Other Files:**
```
services/symbols/
├── go.mod                    # Service module definition
├── Makefile                  # Service commands
├── buf.yaml                  # Service buf config
├── buf.gen.yaml             # Service buf generation
├── Dockerfile               # Production container
└── Dockerfile.debug         # Debug container
```
</service_structure>

<quick_paths>
## Quick Path Reference

**To find a specific component, use these patterns:**

| Component Type | Path Pattern |
|---------------|--------------|
| Proto definition | `api/{service}/v1/{service}.proto` |
| Generated contracts | `contracts/{service}/v1/` |
| Platform utilities | `platform/{package}/` |
| Service entry point | `services/{service}/cmd/{service}/main.go` |
| Wire definitions | `services/{service}/cmd/{service}/wire.go` |
| Business interfaces | `services/{service}/internal/biz/interfaces.go` |
| Business models | `services/{service}/internal/biz/models.go` |
| Use case logic | `services/{service}/internal/biz/{entity}.go` |
| Use case tests | `services/{service}/internal/biz/{entity}_test.go` |
| GORM entities | `services/{service}/internal/data/model/{entity}.go` |
| Repository impl | `services/{service}/internal/data/repo/{entity}.go` |
| gRPC handlers | `services/{service}/internal/service/{entity}.go` |
| DTO mappers | `services/{service}/internal/service/mapper.go` |
| Server config | `services/{service}/internal/server/{grpc\|http}.go` |
| Config schema | `services/{service}/internal/conf/conf.proto` |
| Runtime config | `services/{service}/configs/config.yaml` |
| Service Makefile | `services/{service}/Makefile` |
| Built binary | `services/{service}/bin/{service}` |

**Currently active service:** `symbols`
</quick_paths>

<path_conventions>
## Naming Conventions

**Files:**
- Use snake_case: `symbols.go`, `symbols_test.go`
- Tests co-located: `{name}_test.go` next to `{name}.go`
- Wire generated: `wire_gen.go` (never manually edit)

**Directories:**
- Lowercase singular: `biz/`, `data/`, `service/`, `server/`, `conf/`
- Plurals for collections: `configs/`, `services/`
- Nested by domain: `data/model/`, `data/repo/`

**Import Paths:**
- Contracts: `brizy-go-services/contracts/{service}/v1`
- Platform: `brizy-go-platform/{package}`
- Internal: Cannot be imported from outside service
</path_conventions>

<usage_examples>
## Common Usage Scenarios

**1. Where is the business logic for symbols?**
```
services/symbols/internal/biz/symbols.go
```

**2. Where do I define new repository methods?**
Interface: `services/symbols/internal/biz/interfaces.go`
Implementation: `services/symbols/internal/data/repo/symbol.go`

**3. Where do I add new gRPC methods?**
Proto: `api/symbols/v1/symbols.proto`
Handler: `services/symbols/internal/service/symbols.go`

**4. Where is the database entity?**
```
services/symbols/internal/data/model/symbol.go
```

**5. Where do I add business validation?**
```
services/symbols/internal/biz/validator.go
```

**6. Where is the service configuration?**
Schema: `services/symbols/internal/conf/conf.proto`
Values: `services/symbols/configs/config.yaml`

**7. Where are pagination utilities?**
```
platform/pagination/pagination.go
```

**8. Where do I run tests from?**
```
cd services/symbols
make test
```

**9. Where is the generated Wire code?**
```
services/symbols/cmd/symbols/wire_gen.go
```
(Auto-generated by `make generate`)

**10. Where are proto contracts generated?**
```
contracts/symbols/v1/
```
(Generated by `make contracts-generate` from root)
</usage_examples>

<layer_mapping>
## Clean Architecture Layer Paths

**Service Layer** (external interface):
- Handlers: `services/{service}/internal/service/{entity}.go`
- Mappers: `services/{service}/internal/service/mapper.go`
- Server setup: `services/{service}/internal/server/`

**Business Layer** (use cases):
- Interfaces: `services/{service}/internal/biz/interfaces.go`
- Models: `services/{service}/internal/biz/models.go`
- Logic: `services/{service}/internal/biz/{entity}.go`
- Validation: `services/{service}/internal/biz/validator.go`

**Data Layer** (persistence):
- Setup: `services/{service}/internal/data/data.go`
- Entities: `services/{service}/internal/data/model/{entity}.go`
- Repos: `services/{service}/internal/data/repo/{entity}.go`

**Dependency Flow:** Service → Biz → Data
</layer_mapping>

<working_directory>
## Current Working Directory Context

When working with files, remember:

**From root** (`/Users/alex/GolandProjects/brizy-go-services/`):
- Contract operations: `make contracts-*`
- Access all modules via workspace

**From service** (`cd services/symbols`):
- Service operations: `make generate`, `make test`, `make build`
- Relative paths start from service root

**Absolute paths always work:**
```
/Users/alex/GolandProjects/brizy-go-services/services/symbols/internal/biz/symbols.go
```

**Relative paths depend on current directory:**
```
# From service directory
internal/biz/symbols.go

# From root
services/symbols/internal/biz/symbols.go
```
</working_directory>

<success_criteria>
You can successfully use this skill when you:
- Instantly know where to find any component type
- Can construct correct paths for new files following conventions
- Understand the relationship between layers and their locations
- Know which directory to run commands from
- Can navigate between workspace modules correctly
</success_criteria>