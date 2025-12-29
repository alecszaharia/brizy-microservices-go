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
    Update(context.Context, *Symbol) (*Symbol, error)
    Delete(context.Context, uint64) error
    ListSymbols(context.Context, *ListSymbolsOptions) ([]*Symbol, *pagination.PaginationMeta, error)
}
```

The data layer implements these interfaces:

```go
// internal/data/repo/symbol.go
type symbolRepo struct {
    db *gorm.DB
}

func NewSymbolRepo(db *gorm.DB) biz.SymbolRepo {
    return &symbolRepo{db: db}
}

func (r *symbolRepo) Create(ctx context.Context, s *biz.Symbol) (*biz.Symbol, error) {
    // implementation
}
```

### Wire for Dependency Injection

Wire eliminates manual dependency wiring:

1. **Define dependencies** in `cmd/{service}/wire.go`:
   ```go
   //go:build wireinject
   // +build wireinject

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

2. **Generate code** with `make generate` (runs `wire` with GOWORK=off)
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
type SymbolUseCase interface {
    CreateSymbol(ctx context.Context, s *Symbol) (*Symbol, error)
    GetSymbol(ctx context.Context, id uint64) (*Symbol, error)
    UpdateSymbol(ctx context.Context, s *Symbol) (*Symbol, error)
    DeleteSymbol(ctx context.Context, id uint64) error
    ListSymbols(ctx context.Context, opts *ListSymbolsOptions) ([]*Symbol, *pagination.PaginationMeta, error)
}

type symbolUseCase struct {
    repo      SymbolRepo
    validator SymbolValidator
    log       *log.Helper
}

func NewSymbolUseCase(repo SymbolRepo, validator SymbolValidator, logger log.Logger) SymbolUseCase {
    return &symbolUseCase{
        repo:      repo,
        validator: validator,
        log:       log.NewHelper(logger),
    }
}

func (uc *symbolUseCase) CreateSymbol(ctx context.Context, s *Symbol) (*Symbol, error) {
    // 1. Validate business rules
    if err := uc.validator.ValidateCreate(s); err != nil {
        return nil, err
    }
    
    // 2. Apply business logic (if any)
    // ...
    
    // 3. Persist via repository
    return uc.repo.Create(ctx, s)
}
```

## Service Layer Pattern (Handlers)

Service layer maps between transport (gRPC/HTTP) and business logic:

```go
// internal/service/symbols.go
type SymbolService struct {
    pb.UnimplementedSymbolsServiceServer
    uc     biz.SymbolUseCase
    mapper *SymbolMapper
}

func NewSymbolService(uc biz.SymbolUseCase, mapper *SymbolMapper) *SymbolService {
    return &SymbolService{
        uc:     uc,
        mapper: mapper,
    }
}

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

1. **Transport validation** (proto files with protoc-gen-validate):
   ```protobuf
   message CreateSymbolRequest {
       string name = 1 [(validate.rules).string = {
           min_len: 1,
           max_len: 255
       }];
       uint64 project_id = 2 [(validate.rules).uint64 = {gt: 0}];
   }
   ```

2. **Business validation** (biz layer):
   ```go
   // internal/biz/validator.go
   type SymbolValidator interface {
       ValidateCreate(s *Symbol) error
       ValidateUpdate(s *Symbol) error
   }

   func (v *symbolValidator) ValidateCreate(s *Symbol) error {
       // Business rules validation
       // e.g., uniqueness checks, business constraints
       if s.Name == "" {
           return errors.New("symbol name is required")
       }
       // Check for duplicates, etc.
       return nil
   }
   ```

## Error Handling Strategy

### Business Errors
Define domain errors in biz layer:

```go
// internal/biz/errors.go
var (
    ErrSymbolNotFound       = errors.New("symbol not found")
    ErrSymbolAlreadyExists  = errors.New("symbol already exists")
    ErrInvalidSymbolData    = errors.New("invalid symbol data")
)
```

### Error Propagation
- Return errors up the stack
- Wrap errors with context when needed using `fmt.Errorf("context: %w", err)`
- Service layer can map to gRPC status codes if needed

## Context Usage Patterns

### Request-Scoped Data
Use context for:
- Request ID (via platform/middleware)
- Authentication/authorization
- Request deadlines/timeouts
- Tracing/logging correlation

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
    if tx.Error != nil {
        return tx.Error
    }
    
    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}

// Usage in repository
func (r *symbolRepo) ComplexOperation(ctx context.Context, data *Data) error {
    return common.WithTransaction(r.db, func(tx *gorm.DB) error {
        // Multiple database operations
        if err := tx.Create(&entity1).Error; err != nil {
            return err
        }
        if err := tx.Create(&entity2).Error; err != nil {
            return err
        }
        return nil
    })
}
```

## API Design Patterns

### RESTful Mapping
gRPC methods map to HTTP via google.api.http annotations:

```protobuf
service SymbolsService {
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
    
    rpc UpdateSymbol(UpdateSymbolRequest) returns (UpdateSymbolResponse) {
        option (google.api.http) = {
            put: "/v1/symbols/{id}"
            body: "*"
        };
    }
    
    rpc DeleteSymbol(DeleteSymbolRequest) returns (DeleteSymbolResponse) {
        option (google.api.http) = {
            delete: "/v1/symbols/{id}"
        };
    }
    
    rpc ListSymbols(ListSymbolsRequest) returns (ListSymbolsResponse) {
        option (google.api.http) = {
            get: "/v1/projects/{project_id}/symbols"
        };
    }
}
```

This generates both gRPC and HTTP handlers automatically.

### Pagination Pattern
Use platform/pagination for consistent pagination:

```go
import "brizy-go-platform/pagination"

// In biz layer
type ListSymbolsOptions struct {
    ProjectID uint64
    Offset    uint32
    Limit     uint32
}

func (uc *symbolUseCase) ListSymbols(ctx context.Context, opts *ListSymbolsOptions) ([]*Symbol, *pagination.PaginationMeta, error) {
    return uc.repo.ListSymbols(ctx, opts)
}

// In data layer (repo)
func (r *symbolRepo) ListSymbols(ctx context.Context, opts *biz.ListSymbolsOptions) ([]*biz.Symbol, *pagination.PaginationMeta, error) {
    var symbols []model.Symbol
    var total int64
    
    query := r.db.Model(&model.Symbol{}).Where("project_id = ?", opts.ProjectID)
    
    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, nil, err
    }
    
    // Apply pagination
    if err := query.Offset(int(opts.Offset)).Limit(int(opts.Limit)).Find(&symbols).Error; err != nil {
        return nil, nil, err
    }
    
    // Create pagination metadata
    meta := pagination.NewPaginationMeta(total, opts.Offset, opts.Limit)
    
    return convertToBusinessModels(symbols), meta, nil
}
```

## Configuration Pattern

### Protobuf-Based Config
Define configuration schema in proto:

```protobuf
// internal/conf/conf.proto
syntax = "proto3";

package conf;

message Bootstrap {
    Server server = 1;
    Data data = 2;
}

message Server {
    message HTTP {
        string network = 1;
        string addr = 2;
        int64 timeout = 3;
    }
    message GRPC {
        string network = 1;
        string addr = 2;
        int64 timeout = 3;
    }
    HTTP http = 1;
    GRPC grpc = 2;
}

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
KRATOS_DATA_DATABASE_SOURCE="host=localhost dbname=mydb" ./bin/service
```

## Testing Patterns

### Mock Repositories
Use testify/mock for repository mocks:

```go
// internal/biz/symbols_test.go
type MockSymbolRepo struct {
    mock.Mock
}

func (m *MockSymbolRepo) Create(ctx context.Context, s *biz.Symbol) (*biz.Symbol, error) {
    args := m.Called(ctx, s)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*biz.Symbol), args.Error(1)
}

func (m *MockSymbolRepo) FindByID(ctx context.Context, id uint64) (*biz.Symbol, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*biz.Symbol), args.Error(1)
}
```

### Table-Driven Tests
```go
func TestSymbolUseCase_CreateSymbol(t *testing.T) {
    tests := []struct {
        name    string
        input   *biz.Symbol
        want    *biz.Symbol
        wantErr bool
        setup   func(*MockSymbolRepo, *MockSymbolValidator)
    }{
        {
            name:  "successful creation",
            input: &biz.Symbol{Name: "test", ProjectID: 1},
            want:  &biz.Symbol{ID: 1, Name: "test", ProjectID: 1},
            wantErr: false,
            setup: func(repo *MockSymbolRepo, validator *MockSymbolValidator) {
                validator.On("ValidateCreate", mock.Anything).Return(nil)
                repo.On("Create", mock.Anything, mock.Anything).Return(&biz.Symbol{ID: 1, Name: "test", ProjectID: 1}, nil)
            },
        },
        {
            name:  "validation error",
            input: &biz.Symbol{Name: "", ProjectID: 1},
            want:  nil,
            wantErr: true,
            setup: func(repo *MockSymbolRepo, validator *MockSymbolValidator) {
                validator.On("ValidateCreate", mock.Anything).Return(errors.New("name required"))
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockSymbolRepo)
            mockValidator := new(MockSymbolValidator)
            tt.setup(mockRepo, mockValidator)
            
            uc := biz.NewSymbolUseCase(mockRepo, mockValidator, log.DefaultLogger)
            got, err := uc.CreateSymbol(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
            
            mockRepo.AssertExpectations(t)
            mockValidator.AssertExpectations(t)
        })
    }
}
```
