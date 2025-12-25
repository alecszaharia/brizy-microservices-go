# Design Patterns and Guidelines

## Clean Architecture Principles

### Dependency Rule
Dependencies point inward (from outer layers to inner layers):
```
Service (outer) → Biz (middle) → Data (inner)
```

- **Service layer** depends on **Biz layer**
- **Biz layer** defines interfaces, **Data layer** implements them
- **NO** reverse dependencies allowed
- **NO** layer skipping (service cannot directly access data)

### Interface-Based Design

The business layer (biz) defines interfaces:

```go
// internal/biz/interfaces.go
type SymbolRepo interface {
    Create(context.Context, *Symbol) (*Symbol, error)
    FindByID(context.Context, uint64) (*Symbol, error)
    // ...
}
```

The data layer implements these interfaces:

```go
// internal/data/repo/symbol.go
type symbolRepo struct {
    db *gorm.DB
}

func (r *symbolRepo) Create(ctx context.Context, s *Symbol) (*Symbol, error) {
    // implementation
}
```

### Wire for Dependency Injection

Wire eliminates manual dependency wiring:

1. **Define dependencies** in `cmd/{service}/wire.go`:
   ```go
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

2. **Generate code** with `make generate`
3. **Never manually edit** `wire_gen.go`

## Repository Pattern

All database access goes through repositories:

### Business Layer (defines interface)
```go
// internal/biz/interfaces.go
type SymbolRepo interface {
    Create(context.Context, *Symbol) (*Symbol, error)
    Update(context.Context, *Symbol) (*Symbol, error)
    FindByID(context.Context, uint64) (*Symbol, error)
    Delete(context.Context, uint64) error
    ListSymbols(context.Context, *ListSymbolsOptions) ([]*Symbol, *pagination.PaginationMeta, error)
}
```

### Data Layer (implements interface)
```go
// internal/data/repo/symbol.go
type symbolRepo struct {
    db *gorm.DB
}

func NewSymbolRepo(db *gorm.DB) biz.SymbolRepo {
    return &symbolRepo{db: db}
}
```

## Use Case Pattern

Business logic is encapsulated in use cases:

```go
// internal/biz/symbols.go
type symbolUseCase struct {
    repo      SymbolRepo
    validator SymbolValidator
    log       *log.Helper
}

func (uc *symbolUseCase) CreateSymbol(ctx context.Context, s *Symbol) (*Symbol, error) {
    // 1. Validate business rules
    if err := uc.validator.ValidateCreate(s); err != nil {
        return nil, err
    }
    
    // 2. Apply business logic
    // ...
    
    // 3. Persist via repository
    return uc.repo.Create(ctx, s)
}
```

## Service Layer Pattern (Handlers)

Service layer maps between transport (gRPC/HTTP) and business logic:

```go
// internal/service/symbols.go
func (s *SymbolService) CreateSymbol(ctx context.Context, req *pb.CreateSymbolRequest) (*pb.CreateSymbolResponse, error) {
    // 1. Map DTO to business model
    symbol := s.mapper.RequestToSymbol(req)
    
    // 2. Call use case
    created, err := s.uc.CreateSymbol(ctx, symbol)
    if err != nil {
        return nil, err
    }
    
    // 3. Map business model to DTO
    return &pb.CreateSymbolResponse{
        Symbol: s.mapper.SymbolToProto(created),
    }, nil
}
```

## Validation Strategy

### Two-Level Validation

1. **Transport validation** (proto files):
   ```protobuf
   message CreateSymbolRequest {
       string name = 1 [(validate.rules).string = {
           min_len: 1,
           max_len: 255
       }];
   }
   ```

2. **Business validation** (biz layer):
   ```go
   func (v *symbolValidator) ValidateCreate(s *Symbol) error {
       // Business rules validation
       // e.g., uniqueness, business constraints
   }
   ```

## Error Handling Strategy

### Business Errors
Define domain errors in biz layer:

```go
// internal/biz/errors.go
var (
    ErrSymbolNotFound = errors.New("symbol not found")
    ErrSymbolExists   = errors.New("symbol already exists")
)
```

### Error Propagation
- Return errors up the stack
- Wrap errors with context when needed
- Service layer can map to gRPC status codes

## Context Usage Patterns

### Request-Scoped Data
Use context for:
- Request ID (via platform middleware)
- Authentication/authorization
- Request deadlines/timeouts

```go
func (uc *symbolUseCase) GetSymbol(ctx context.Context, id uint64) (*Symbol, error) {
    // Context automatically carries request ID for logging
    uc.log.WithContext(ctx).Infof("Getting symbol %d", id)
    return uc.repo.FindByID(ctx, id)
}
```

## Transaction Management

For operations requiring multiple repository calls:

```go
// internal/data/common/transaction.go
func WithTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
    tx := db.Begin()
    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit().Error
}
```

## API Design Patterns

### RESTful Mapping
gRPC methods map to HTTP via annotations:

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

### Pagination Pattern
Use platform/pagination for consistent pagination:

```go
opts := &biz.ListSymbolsOptions{
    Offset: req.Offset,
    Limit:  req.Limit,
}
symbols, meta, err := uc.ListSymbols(ctx, opts)
```

## Configuration Pattern

### Protobuf-Based Config
Define configuration schema in proto:

```protobuf
// internal/conf/conf.proto
message Data {
    message Database {
        string driver = 1;
        string source = 2;
    }
    Database database = 1;
}
```

### Environment Overrides
Kratos supports env variable overrides with `KRATOS_` prefix:
```bash
KRATOS_DATA_DATABASE_SOURCE="host=localhost" ./bin/service
```

## Testing Patterns

### Mock Repositories
Use testify/mock for repository mocks:

```go
type MockSymbolRepo struct {
    mock.Mock
}

func (m *MockSymbolRepo) Create(ctx context.Context, s *biz.Symbol) (*biz.Symbol, error) {
    args := m.Called(ctx, s)
    return args.Get(0).(*biz.Symbol), args.Error(1)
}
```

### Table-Driven Tests
```go
func TestSymbolUseCase_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   *Symbol
        want    *Symbol
        wantErr bool
    }{
        {
            name: "valid symbol",
            input: &Symbol{Name: "test"},
            want: &Symbol{ID: 1, Name: "test"},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```
