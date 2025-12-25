# Code Style and Conventions

## General Go Style

Follow standard Go conventions:
- `gofmt` formatting (enforced by editor/CI)
- Exported functions/types start with uppercase
- Private functions/types start with lowercase
- Use meaningful variable names (avoid single letters except in loops)

## Naming Conventions

### Files
- Use snake_case for file names: `symbols.go`, `symbols_test.go`
- Test files: `{name}_test.go` co-located with implementation
- Group related functionality in same file

### Packages
- Package names are lowercase, single word: `biz`, `data`, `service`, `repo`
- Package imports use full module path: `brizy-go-platform/pagination`

### Types and Interfaces
- Interface names: `{Entity}UseCase`, `{Entity}Repo`
  - Example: `SymbolUseCase`, `SymbolRepo`
- Struct names: `{Entity}`, `{Entity}Options`
  - Example: `Symbol`, `ListSymbolsOptions`
- Use descriptive names for business models

### Functions and Methods
- Use verb prefixes: `Get`, `Create`, `Update`, `Delete`, `List`, `Find`
- Repository methods: `Create`, `Update`, `FindByID`, `Delete`, `List{Entity}s`
- Use case methods: match repository but may add validation/business logic
- Service handlers: match proto RPC method names

## Comments and Documentation

### Required Comments
- All exported types, functions, and methods must have comments
- Comments start with the name of the item being documented
- Example:
  ```go
  // SymbolUseCase defines the business logic for symbol operations.
  type SymbolUseCase interface {
      // GetSymbol retrieves a Symbol by its ID.
      GetSymbol(ctx context.Context, id uint64) (*Symbol, error)
  }
  ```

### Optional Comments
- Internal implementation details (only if non-obvious)
- Complex algorithms or business rules
- TODO markers for future work

## Architecture Patterns

### Clean Architecture Layers

**Layer separation is strict:**

1. **Service Layer** (`internal/service/`)
   - Handles gRPC/HTTP requests
   - Validates input (via protoc-gen-validate)
   - Calls business logic (use cases)
   - Maps between DTOs and business models
   - Returns gRPC/HTTP responses

2. **Business Layer** (`internal/biz/`)
   - Defines business models and interfaces
   - Implements use cases (business logic)
   - Validates business rules
   - Depends on repository interfaces (not implementations)
   - NO database, NO external services, NO frameworks

3. **Data Layer** (`internal/data/`)
   - Implements repository interfaces from biz layer
   - GORM entities and database operations
   - Transaction management
   - External service integrations

### Dependency Injection with Wire

- Define dependencies in `cmd/{service}/wire.go`
- Run `make generate` to create `wire_gen.go`
- Each layer exports a `ProviderSet` in its main file
- Never manually edit `wire_gen.go`

### Error Handling

- Return errors from all fallible operations
- Use context.Context as first parameter
- Wrap errors with context when propagating up
- Define business errors in `biz/errors.go`

### Context Usage

- Always pass `context.Context` as first parameter
- Use context for request-scoped data (request ID, auth, etc.)
- Use platform middleware for request ID propagation

## Testing Patterns

### Unit Tests
- Co-locate with implementation: `symbols.go` → `symbols_test.go`
- Use `testify/assert` for assertions
- Use `testify/mock` for mocking repository interfaces
- Table-driven tests for multiple scenarios
- Test file structure:
  ```go
  func TestSymbolUseCase_Create(t *testing.T) {
      tests := []struct {
          name    string
          input   *Symbol
          want    *Symbol
          wantErr bool
      }{
          // test cases
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              // test implementation
          })
      }
  }
  ```

## Protocol Buffers Style

### File Organization
- One service per proto file
- Define in `api/{service}/v1/{service}.proto`
- Use semantic versioning in package path

### Message Naming
- Request: `{Action}{Entity}Request`
- Response: `{Action}{Entity}Response`
- Example: `CreateSymbolRequest`, `CreateSymbolResponse`

### RPC Naming
- Use verb + noun: `CreateSymbol`, `GetSymbol`, `ListSymbols`
- Include HTTP annotations for REST mapping
- Add validation rules with protoc-gen-validate

## Database Conventions (GORM)

- Define entities in `internal/data/model/`
- Use struct tags for table/column mapping
- Implement soft deletes with `gorm.DeletedAt`
- Repository pattern for all database access
- Use transactions for multi-step operations
