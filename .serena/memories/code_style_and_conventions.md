# Code Style and Conventions

## General Go Style

Follow standard Go conventions:
- `gofmt` formatting (enforced by editor/CI)
- Exported (public) functions/types start with uppercase
- Private (unexported) functions/types start with lowercase
- Use meaningful variable names (avoid single letters except in loops/short scopes)
- Prefer simplicity over cleverness

## Naming Conventions

### Files
- Use snake_case for file names: `symbols.go`, `symbols_test.go`
- Test files: `{name}_test.go` co-located with implementation
- Group related functionality in same file
- Main files per layer:
  - `biz.go` - exports `ProviderSet` for biz layer
  - `data.go` - exports database setup and `ProviderSet` for data layer
  - `service.go` - exports service struct and `ProviderSet` for service layer
  - `server.go` - exports server setup and `ProviderSet` for server layer

### Packages
- Package names are lowercase, single word: `biz`, `data`, `service`, `repo`, `model`
- Avoid package names like `utils`, `helpers`, `common` (be specific)
- Package imports use full module path:
  - `brizy-go-platform/pagination`
  - `brizy-go-platform/middleware`
  - `contracts/symbols/v1`

### Types and Interfaces
- Interface names: `{Entity}UseCase`, `{Entity}Repo`, `{Entity}Validator`
  - Example: `SymbolUseCase`, `SymbolRepo`, `SymbolValidator`
- Struct names: `{entity}`, `{Entity}Options`, `{Entity}Request`
  - Example: `Symbol`, `ListSymbolsOptions`, `CreateSymbolRequest`
- Use descriptive names for business models
- Avoid stuttering: `symbol.Symbol` not `symbol.SymbolEntity`

### Functions and Methods
- Use verb prefixes: `Get`, `Create`, `Update`, `Delete`, `List`, `Find`, `New`, `With`
- Repository interface methods:
  - `Create(ctx context.Context, entity *Entity) (*Entity, error)`
  - `Update(ctx context.Context, entity *Entity) (*Entity, error)`
  - `FindByID(ctx context.Context, id uint64) (*Entity, error)`
  - `Delete(ctx context.Context, id uint64) error`
  - `List{Entity}s(ctx context.Context, opts *ListOptions) ([]*Entity, *pagination.PaginationMeta, error)`
- Use case methods: match repository but may add validation/business logic
- Service handlers: match proto RPC method names exactly
- Constructor functions: `New{Type}(deps) *{Type}`

### Variables and Constants
- Use camelCase for variables: `symbolRepo`, `userName`, `projectID`
- Use UPPER_SNAKE_CASE for constants (if they're truly constant)
- Prefer single-letter variables for short scopes only (i, j, k in loops)
- Context always named `ctx`
- Logger always named `log` or `logger`

## Comments and Documentation

### Required Comments
- All exported types, functions, methods, and constants must have comments
- Comments start with the name of the item being documented
- Example:
  ```go
  // SymbolUseCase defines the business logic for symbol operations.
  type SymbolUseCase interface {
      // CreateSymbol creates a new Symbol and returns it.
      CreateSymbol(ctx context.Context, s *Symbol) (*Symbol, error)
      
      // GetSymbol retrieves a Symbol by its ID.
      GetSymbol(ctx context.Context, id uint64) (*Symbol, error)
  }
  ```

### Optional Comments
- Internal implementation details (only if non-obvious)
- Complex algorithms or business rules
- TODO markers for future work (format: `// TODO: description`)
- FIXME markers for known issues (format: `// FIXME: description`)

### Package Comments
- Every package should have a package comment
- Place in a file named `doc.go` or at the top of the main file
- Example:
  ```go
  // Package biz contains business logic and use case implementations.
  package biz
  ```

## Architecture Patterns

### Clean Architecture Layers

**Layer separation is strict:**

1. **Service Layer** (`internal/service/`)
   - Handles gRPC/HTTP requests
   - Validates input (via protoc-gen-validate - automatic)
   - Calls business logic (use cases)
   - Maps between DTOs (protobuf) and business models
   - Returns gRPC/HTTP responses
   - **Dependencies**: `biz` layer only

2. **Business Layer** (`internal/biz/`)
   - Defines business models (`models.go`)
   - Defines repository interfaces (`interfaces.go`)
   - Implements use cases (business logic)
   - Validates business rules (`validator.go`)
   - Defines business errors (`errors.go`)
   - **Dependencies**: NO external frameworks, NO database, NO transport
   - **Only depends on**: Standard library, platform utilities

3. **Data Layer** (`internal/data/`)
   - Implements repository interfaces from biz layer
   - GORM entities (`model/`) - ORM-specific
   - Repository implementations (`repo/`)
   - Database operations, transactions
   - External service integrations
   - **Dependencies**: GORM, database drivers, biz interfaces

### Dependency Injection with Wire

- Define dependencies in `cmd/{service}/wire.go`
- Use `//go:build wireinject` build tag
- Run `make generate` to create `wire_gen.go`
- Each layer exports a `ProviderSet` in its main file (e.g., `biz.go`, `data.go`)
- **Never manually edit** `wire_gen.go`

Example wire.go structure:
```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/go-kratos/kratos/v2/log"
    "github.com/google/wire"
    "your-service/internal/biz"
    "your-service/internal/conf"
    "your-service/internal/data"
    "your-service/internal/server"
    "your-service/internal/service"
)

func InitApp(*conf.Server, *conf.Data, log.Logger) (*App, error) {
    wire.Build(
        server.ProviderSet,
        service.ProviderSet,
        biz.ProviderSet,
        data.ProviderSet,
        newApp,
    )
    return &App{}, nil
}
```

### Error Handling

- **Always** return errors from fallible operations
- Use `context.Context` as **first parameter** in all methods
- Return error as **last return value**
- Wrap errors with context: `fmt.Errorf("failed to create symbol: %w", err)`
- Define business errors in `biz/errors.go`
- Check errors immediately after function calls
- Don't ignore errors (use `_ = err` if intentionally ignoring)

Example:
```go
symbol, err := uc.repo.Create(ctx, s)
if err != nil {
    return nil, fmt.Errorf("failed to create symbol: %w", err)
}
```

### Context Usage

- **Always** pass `context.Context` as first parameter
- Use context for:
  - Request-scoped data (request ID, auth, user info)
  - Cancellation signals
  - Deadlines/timeouts
  - Tracing/logging correlation
- Use platform middleware for request ID propagation
- Pass context through all layers (service → biz → data)
- Don't store context in structs (pass as parameter)

## Testing Patterns

### Unit Tests
- Co-locate with implementation: `symbols.go` → `symbols_test.go`
- Use `testify/assert` for assertions
- Use `testify/mock` for mocking repository interfaces
- Table-driven tests for multiple scenarios
- Test names: `Test{Type}_{Method}_{Scenario}`
  - Example: `TestSymbolUseCase_CreateSymbol_Success`
  - Example: `TestSymbolUseCase_CreateSymbol_ValidationError`

### Test Structure
```go
func Test{Type}_{Method}(t *testing.T) {
    tests := []struct {
        name    string        // Test case name
        input   *InputType    // Input data
        want    *OutputType   // Expected output
        wantErr bool          // Expect error?
        setup   func(*mocks)  // Setup mocks
    }{
        {
            name: "successful case",
            input: &Input{...},
            want: &Output{...},
            wantErr: false,
            setup: func(m *mocks) {
                m.repo.On("Create", mock.Anything, mock.Anything).Return(&Entity{...}, nil)
            },
        },
        {
            name: "error case",
            input: &Input{...},
            want: nil,
            wantErr: true,
            setup: func(m *mocks) {
                m.repo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mocks := setupMocks()
            if tt.setup != nil {
                tt.setup(mocks)
            }
            
            // Act
            got, err := subject.Method(context.Background(), tt.input)
            
            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, got)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
            
            mocks.AssertExpectations(t)
        })
    }
}
```

## Protocol Buffers Style

### File Organization
- One service per proto file
- Define in `api/{service}/v1/{service}.proto`
- Use semantic versioning in package path (v1, v2, etc.)
- Optional: `error_reason.proto` for service-specific error enums

### Message Naming
- Request: `{Action}{Entity}Request`
- Response: `{Action}{Entity}Response`
- Entity: `{Entity}` (singular)
- List wrapper: `{Entity}List` or use repeated in response
- Example:
  - `CreateSymbolRequest`, `CreateSymbolResponse`
  - `GetSymbolRequest`, `GetSymbolResponse`
  - `Symbol` (the entity message)

### RPC Naming
- Use verb + noun: `CreateSymbol`, `GetSymbol`, `ListSymbols`, `UpdateSymbol`, `DeleteSymbol`
- Include HTTP annotations for REST mapping
- Add validation rules with protoc-gen-validate
- Use semantic HTTP methods:
  - POST for Create
  - GET for Read
  - PUT for Update
  - DELETE for Delete

Example:
```protobuf
rpc CreateSymbol(CreateSymbolRequest) returns (CreateSymbolResponse) {
    option (google.api.http) = {
        post: "/v1/symbols"
        body: "*"
    };
}

rpc GetSymbol(GetSymbolRequest) returns (GetSymbolResponse) {
    option (google.api.http) = {
        get: "/v1/symbols/{id}"
    };
}
```

### Validation Rules
Use protoc-gen-validate for transport-level validation:
```protobuf
message CreateSymbolRequest {
    string name = 1 [(validate.rules).string = {
        min_len: 1,
        max_len: 255
    }];
    uint64 project_id = 2 [(validate.rules).uint64 = {gt: 0}];
}
```

## Database Conventions (GORM)

### Entity Definitions
- Define entities in `internal/data/model/`
- Use struct tags for table/column mapping
- Implement soft deletes with `gorm.DeletedAt`
- Use `gorm.Model` or define ID, CreatedAt, UpdatedAt, DeletedAt manually

Example:
```go
// internal/data/model/symbol.go
package model

import "gorm.io/gorm"

type Symbol struct {
    ID              uint64 `gorm:"primaryKey;autoIncrement"`
    ProjectID       uint64 `gorm:"index;not null"`
    UID             string `gorm:"size:255;not null"`
    Label           string `gorm:"size:255;not null"`
    ClassName       string `gorm:"size:255;not null"`
    ComponentTarget string `gorm:"size:255;not null"`
    Version         uint32 `gorm:"not null"`
    CreatedAt       int64  `gorm:"autoCreateTime"`
    UpdatedAt       int64  `gorm:"autoUpdateTime"`
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Symbol) TableName() string {
    return "symbols"
}
```

### Repository Pattern
- All database access through repositories
- Repository implements interface from `biz/interfaces.go`
- Use GORM query builder
- Handle errors properly
- Use transactions for multi-step operations

## Import Organization

Group imports in this order:
1. Standard library
2. External dependencies
3. Internal packages (same project)

Example:
```go
import (
    "context"
    "fmt"
    
    "github.com/go-kratos/kratos/v2/log"
    "gorm.io/gorm"
    
    "your-service/internal/biz"
    "brizy-go-platform/pagination"
)
```

## Key Principles

1. **Simplicity**: Keep code simple and readable
2. **Testability**: Write testable code (use interfaces)
3. **Separation of Concerns**: Respect layer boundaries
4. **Consistency**: Follow existing patterns in the codebase
5. **Documentation**: Document public APIs
6. **Error Handling**: Handle all errors, don't ignore
7. **Context Propagation**: Pass context through all layers
