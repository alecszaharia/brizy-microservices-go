---
skill_name: path-finder
description: Provides correct file paths for brizy-go-services monorepo based on Clean Architecture conventions and workspace structure. Use when you need to reference or locate files in the codebase.
tags: [paths, file-structure, clean-architecture, monorepo, workspace]
version: 1.0.0
last_updated: 2025-12-28
---

# Path Finder Skill

## Purpose

Navigate the brizy-go-services monorepo file structure following Clean Architecture patterns and Go workspace conventions.

## Workspace Structure

```
brizy-go-services/
├── go.work                          # Go workspace definition
├── go.work.sum                      # Workspace checksums
├── contracts/                       # Shared protobuf module
│   ├── go.mod
│   └── gen/                         # Generated code
│       └── {service}/
│           └── v1/
│               ├── {service}.pb.go
│               ├── {service}_grpc.pb.go
│               └── {service}.swagger.json
├── platform/                        # Shared utilities module
│   ├── go.mod
│   ├── middleware/
│   ├── pagination/
│   └── adapters/
└── services/                        # Microservices
    └── {service}/                   # Individual service
        ├── go.mod
        ├── Makefile
        ├── cmd/
        │   └── {service}/
        │       ├── main.go
        │       ├── wire.go
        │       └── wire_gen.go
        ├── internal/
        │   ├── biz/                 # Business logic layer
        │   │   ├── interfaces.go
        │   │   ├── models.go
        │   │   ├── {entity}.go
        │   │   └── {entity}_test.go
        │   ├── data/                # Data access layer
        │   │   ├── data.go
        │   │   ├── model/
        │   │   │   └── {entity}.go
        │   │   ├── repo/
        │   │   │   ├── {entity}.go
        │   │   │   └── {entity}_test.go
        │   │   └── common/
        │   │       └── transaction.go
        │   ├── service/             # Service layer (handlers)
        │   │   ├── service.go
        │   │   ├── {entity}.go
        │   │   ├── {entity}_test.go
        │   │   ├── mapper.go
        │   │   └── mapper_test.go
        │   ├── server/              # Server setup
        │   │   ├── grpc.go
        │   │   └── http.go
        │   └── conf/                # Configuration
        │       └── conf.proto
        └── configs/
            └── config.yaml
```

## Layer-Specific Paths

### Business Layer (biz)

**Location**: \`services/{service}/internal/biz/\`

**File Patterns**:
- \`interfaces.go\` - Repository interface definitions
- \`models.go\` - Business domain models and DTOs
- \`{entity}.go\` - Use case implementations
- \`{entity}_test.go\` - Use case tests
- \`errors.go\` - Business error definitions
- \`validator.go\` - Validation logic

**Import Path**: \`{service}/internal/biz\`

**Example**:
```go
// services/symbols/internal/biz/symbols.go
package biz

import "context"

type SymbolUseCase interface {
    GetSymbol(ctx context.Context, id uint64) (*Symbol, error)
}
```

### Data Layer (data)

**Location**: \`services/{service}/internal/data/\`

**Subdirectories**:
- \`model/\` - GORM entities (database models)
- \`repo/\` - Repository implementations
- \`common/\` - Shared data layer utilities

**File Patterns**:
- \`data.go\` - Database setup and initialization
- \`model/{entity}.go\` - GORM entity definition
- \`repo/{entity}.go\` - Repository implementation
- \`repo/{entity}_test.go\` - Repository tests
- \`common/transaction.go\` - Transaction utilities

**Import Paths**:
- GORM models: \`{service}/internal/data/model\`
- Repositories: \`{service}/internal/data/repo\`
- Common utilities: \`{service}/internal/data/common\`

**Example**:
```go
// services/symbols/internal/data/repo/symbol.go
package repo

import (
    "symbols/internal/biz"
    "symbols/internal/data/model"
)
```

### Service Layer (service)

**Location**: \`services/{service}/internal/service/\`

**File Patterns**:
- \`service.go\` - Service struct definition
- \`{entity}.go\` - gRPC/HTTP handler implementations
- \`{entity}_test.go\` - Service handler tests
- \`mapper.go\` - DTO ↔ Domain model conversions
- \`mapper_test.go\` - Mapper tests
- \`errors.go\` - Service error mapping

**Import Path**: \`{service}/internal/service\`

**Example**:
```go
// services/symbols/internal/service/symbols.go
package service

import (
    v1 "contracts/gen/symbols/v1"
    "symbols/internal/biz"
)
```

### Server Layer (server)

**Location**: \`services/{service}/internal/server/\`

**Files**:
- \`grpc.go\` - gRPC server setup
- \`http.go\` - HTTP server setup (Kratos bindings)

**Import Path**: \`{service}/internal/server\`

### Configuration

**Location**: \`services/{service}/internal/conf/\`

**Files**:
- \`conf.proto\` - Configuration protobuf schema

**Runtime Config**: \`services/{service}/configs/config.yaml\`

## Shared Modules

### Contracts (Protobuf)

**Location**: \`contracts/gen/{service}/v1/\`

**Generated Files**:
- \`{service}.pb.go\` - Protobuf messages
- \`{service}_grpc.pb.go\` - gRPC service definitions
- \`{service}_http.pb.go\` - Kratos HTTP bindings
- \`{service}.swagger.json\` - OpenAPI spec

**Import Path**: \`contracts/gen/{service}/v1\`

**Example**:
```go
import v1 "contracts/gen/symbols/v1"
```

### Platform Utilities

**Location**: \`platform/{package}/\`

**Packages**:
- \`platform/middleware\` - Request ID middleware
- \`platform/pagination\` - Pagination utilities
- \`platform/adapters\` - Common adapters

**Import Examples**:
```go
import "platform/pagination"
import "platform/middleware"
```

## Path Resolution Rules

### 1. Service-Specific Code

Pattern: \`services/{service}/internal/{layer}/{file}.go\`

Examples:
- Business logic: \`services/symbols/internal/biz/symbols.go\`
- Repository: \`services/symbols/internal/data/repo/symbol.go\`
- GORM model: \`services/symbols/internal/data/model/symbol.go\`
- Service handler: \`services/symbols/internal/service/symbols.go\`

### 2. Test Files

Pattern: \`{same_path_as_implementation}_test.go\`

Examples:
- \`services/symbols/internal/biz/symbols_test.go\`
- \`services/symbols/internal/data/repo/symbol_test.go\`
- \`services/symbols/internal/service/symbols_test.go\`

### 3. Proto Definitions

**Source**: \`api/{service}/v1/{service}.proto\`
**Generated**: \`contracts/gen/{service}/v1/\`

### 4. Configuration

**Proto Schema**: \`services/{service}/internal/conf/conf.proto\`
**Runtime Config**: \`services/{service}/configs/config.yaml\`

## Quick Reference Table

| What | Path Pattern | Example |
|------|-------------|---------|
| Use case interface | \`services/{service}/internal/biz/interfaces.go\` | \`services/symbols/internal/biz/interfaces.go\` |
| Use case implementation | \`services/{service}/internal/biz/{entity}.go\` | \`services/symbols/internal/biz/symbols.go\` |
| Business models | \`services/{service}/internal/biz/models.go\` | \`services/symbols/internal/biz/models.go\` |
| GORM entity | \`services/{service}/internal/data/model/{entity}.go\` | \`services/symbols/internal/data/model/symbol.go\` |
| Repository | \`services/{service}/internal/data/repo/{entity}.go\` | \`services/symbols/internal/data/repo/symbol.go\` |
| Service handler | \`services/{service}/internal/service/{entity}.go\` | \`services/symbols/internal/service/symbols.go\` |
| Mapper | \`services/{service}/internal/service/mapper.go\` | \`services/symbols/internal/service/mapper.go\` |
| Wire config | \`services/{service}/cmd/{service}/wire.go\` | \`services/symbols/cmd/symbols/wire.go\` |
| Main entry | \`services/{service}/cmd/{service}/main.go\` | \`services/symbols/cmd/symbols/main.go\` |
| Proto def | \`api/{service}/v1/{service}.proto\` | \`api/symbols/v1/symbols.proto\` |
| Generated proto | \`contracts/gen/{service}/v1/{service}.pb.go\` | \`contracts/gen/symbols/v1/symbols.pb.go\` |

## Import Path Patterns

### Within Same Service

```go
// From service layer to biz layer
import "{service}/internal/biz"

// From repo to model
import "{service}/internal/data/model"

// From repo to biz (for interfaces)
import "{service}/internal/biz"
```

### Cross-Module Imports

```go
// Platform utilities
import "platform/pagination"
import "platform/middleware"

// Generated protobuf
import v1 "contracts/gen/symbols/v1"

// GORM
import "gorm.io/gorm"

// Kratos
import "github.com/go-kratos/kratos/v2/log"
```

## Common Scenarios

### Finding Entity Files

Given entity name \`Symbol\`:
- Business model: \`services/symbols/internal/biz/models.go\` (struct \`Symbol\`)
- Use case: \`services/symbols/internal/biz/symbols.go\`
- GORM entity: \`services/symbols/internal/data/model/symbol.go\` (struct \`Symbol\`)
- Repository: \`services/symbols/internal/data/repo/symbol.go\`
- Service handler: \`services/symbols/internal/service/symbols.go\`

### Finding Tests

Same directory as implementation + \`_test.go\` suffix:
- \`services/symbols/internal/biz/symbols_test.go\`
- \`services/symbols/internal/data/repo/symbol_test.go\`
- \`services/symbols/internal/service/symbols_test.go\`

### Finding Configuration

- Schema: \`services/symbols/internal/conf/conf.proto\`
- Runtime: \`services/symbols/configs/config.yaml\`
- Wire DI: \`services/symbols/cmd/symbols/wire.go\`

## Directory Naming Conventions

- Use **singular** for: \`biz\`, \`data\`, \`service\`, \`server\`, \`conf\`
- Use **plural** for: \`services\`, \`configs\`, \`contracts\`
- Use **snake_case** for: multi-word directories (e.g., \`symbol_data\`)
- Use **lowercase** throughout

## File Naming Conventions

- Use **singular entity name**: \`symbol.go\` (not \`symbols.go\`) for model/repo
- Use **plural entity name**: \`symbols.go\` for use case/service (matches proto)
- Use **snake_case**: \`symbol_data.go\`
- Tests: \`{filename}_test.go\`

