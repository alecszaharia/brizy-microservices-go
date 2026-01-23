# Use Case Pattern

<overview>
Use cases are the heart of the business layer. They orchestrate business logic, coordinate between repositories, handle validation, and manage errors.
</overview>

<struct_pattern>
## Use Case Struct

**Location**: `internal/biz/{entity}/usecase.go`

**Standard structure**:
```go
// Package {entity} provides use cases for managing {entities}.
package {entity}

type useCase struct {
	repo      domain.{Entity}Repo           // Repository interface
	pub       domain.{Entity}EventPublisher // Event publisher (optional)
	log       *log.Helper                   // Structured logger
	validator *validator.Validate           // Validator instance
	tm        common.Transaction            // Transaction manager
}
```

**Field naming**:
- `repo` - single repository (lowercase)
- `pub` - event publisher (optional)
- `log` - always `*log.Helper`
- `validator` - always `*validator.Validate`
- `tm` - transaction manager

**Struct naming**: `useCase` (camelCase, unexported, generic name)
**Package naming**: `{entity}` (singular, e.g., `package symbol`)
</struct_pattern>

<constructor_pattern>
## Constructor Function

**Standard signature** (with event publishing):
```go
// NewUseCase creates a new {Entity} use case.
func NewUseCase(
	repo domain.{Entity}Repo,
	v *validator.Validate,
	tm common.Transaction,
	pub domain.{Entity}EventPublisher,
	logger log.Logger,
) domain.{Entity}UseCase {  // Returns interface from domain package
	return &useCase{
		repo:      repo,
		validator: v,
		tm:        tm,
		pub:       pub,
		log:       log.NewHelper(logger),
	}
}
```

**Without event publishing**:
```go
func NewUseCase(
	repo domain.{Entity}Repo,
	v *validator.Validate,
	tm common.Transaction,
	logger log.Logger,
) domain.{Entity}UseCase {
	return &useCase{
		repo:      repo,
		validator: v,
		tm:        tm,
		log:       log.NewHelper(logger),
	}
}
```

**Key patterns**:
- Function name: `NewUseCase` (generic, not entity-specific)
- Accept `log.Logger`, wrap as `log.NewHelper(logger)`
- Return **interface type from domain package**: `domain.{Entity}UseCase`
- Parameters in order: repo, validator, tm, pub (optional), logger
- Import domain package: `import "symbols/internal/biz/domain"`
</constructor_pattern>

<method_pattern>
## Use Case Methods

**Method signature**:
```go
func (uc *{entity}UseCase) Create{Entity}(ctx context.Context, e *{Entity}) (*{Entity}, error)
```

**Standard flow**:
1. Validate input with validator
2. Call repository method
3. Handle errors with logging
4. Map repository errors to business errors
5. Return result

**Example**:
```go
func (uc *{entity}UseCase) Create{Entity}(ctx context.Context, e *{Entity}) (*{Entity}, error) {
	// 1. Validate
	if err := uc.validator.Struct(e); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// 2. Call repository
	{entity}, err := uc.repo.Create(ctx, e)
	if err != nil {
		// 3. Log error
		uc.log.WithContext(ctx).Errorf("Create error: %v", err)

		// 4. Map to business error
		if errors.Is(err, ErrDuplicateEntry) {
			return nil, ErrDuplicate{Entity}
		}

		return nil, ErrDatabaseOperation
	}

	// 5. Return result
	return {entity}, nil
}
```
</method_pattern>

<error_handling>
## Error Handling Pattern

**Define errors in domain layer** (`internal/biz/domain/errors.go`):
```go
package domain

import "errors"

// Domain-level errors (returned by business logic layer)
var (
	ErrSymbolNotFound    = errors.New("symbol not found")
	ErrDuplicateSymbol   = errors.New("symbol with this UID already exists")
	ErrInvalidID         = errors.New("invalid symbol ID")
	ErrValidationFailed  = errors.New("validation failed")
	ErrDatabaseOperation = errors.New("database operation failed")
)

// Data layer errors (returned by repository implementations)
var (
	ErrDataNotFound       = errors.New("record not found")
	ErrDataDuplicateEntry = errors.New("duplicate entry")
	ErrDataDatabase       = errors.New("database error")
)
```

**Map data layer errors to domain errors in use case**:
```go
// toDomainError transforms data layer errors to domain errors.
func toDomainError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, domain.ErrDataNotFound) {
		return domain.ErrSymbolNotFound
	}
	if errors.Is(err, domain.ErrDataDuplicateEntry) {
		return domain.ErrDuplicateSymbol
	}
	if errors.Is(err, domain.ErrDataDatabase) {
		return domain.ErrDatabaseOperation
	}

	return err
}
```

**Use the mapper in use case methods**:
```go
symbol, err := uc.repo.FindByID(ctx, id)
if err != nil {
	uc.log.WithContext(ctx).Errorf("Failed to get symbol: %v", err)
	return nil, toDomainError(err)
}
```

**Two-layer error approach**:
- **Repositories** return data errors (`domain.ErrDataNotFound`)
- **Use cases** map to domain errors (`domain.ErrSymbolNotFound`)
- **Service layer** maps domain errors to gRPC/HTTP status codes
</error_handling>

<logging_pattern>
## Logging Pattern

**Use WithContext** for request tracing:
```go
uc.log.WithContext(ctx).Errorf("Create error: %v", err)
uc.log.WithContext(ctx).Infof("Created {entity} %d", {entity}.ID)
```

**Log levels**:
- `Errorf` - Errors that prevent operation completion
- `Warnf` - Recoverable issues or unexpected states
- `Infof` - Important business events
- `Debugf` - Detailed debugging information

**What to log**:
- ✅ Repository errors with context
- ✅ Validation failures (at debug level)
- ✅ Business rule violations
- ❌ Successful operations (creates noise)
- ❌ Sensitive data (passwords, tokens)
</logging_pattern>

<validation_pattern>
## Validation Pattern

**Validate before repository calls**:
```go
// Validate domain model
if err := uc.validator.Struct(e); err != nil {
	return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
}
```

**Validate options/parameters**:
```go
// Validate list options
if err := uc.validator.Struct(options); err != nil {
	return nil, nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
}
```

**Don't re-validate** if data comes from trusted source (e.g., Get after Create).
</validation_pattern>

<transaction_pattern>
## Transaction Pattern

**With event publishing** (most common):
```go
func (uc *useCase) CreateSymbol(ctx context.Context, s *domain.Symbol) (*domain.Symbol, error) {
	// Validate domain model
	if err := uc.validator.Struct(s); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	var symbol *domain.Symbol

	err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		var err error
		symbol, err = uc.repo.Create(ctx, s)
		if err != nil {
			return err  // Rolls back
		}

		// Publish event within transaction
		return uc.pub.PublishSymbolCreated(ctx, symbol)
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Create error: %v", err)
		return nil, toDomainError(err)
	}

	return symbol, nil
}
```

**For multi-repository operations**:
```go
func (uc *useCase) ComplexOperation(ctx context.Context, data *Data) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		// Multiple repository calls
		if err := uc.repo1.Create(ctx, item1); err != nil {
			return err
		}

		if err := uc.repo2.Update(ctx, item2); err != nil {
			return err
		}

		// Publish event last
		return uc.pub.PublishEvent(ctx, data)
	})
}
```

**When to use transactions**:
- Operations with event publishing (ensure consistency)
- Multiple related repository operations
- Operations that must succeed or fail together
- Complex business workflows

**When NOT to use transactions**:
- Single repository call without events
- Read-only operations
- Independent operations

**Critical**: If event publishing fails, the entire transaction rolls back.
</transaction_pattern>

<event_publishing_pattern>
## Event Publishing Pattern

**Define event publisher interface** in `internal/biz/domain/interfaces.go`:
```go
type SymbolEventPublisher interface {
	PublishSymbolCreated(ctx context.Context, symbol *Symbol) error
	PublishSymbolUpdated(ctx context.Context, symbol *Symbol) error
	PublishSymbolDeleted(ctx context.Context, symbol *Symbol) error
}
```

**Inject publisher in use case constructor**:
```go
func NewUseCase(
	repo domain.SymbolRepo,
	v *validator.Validate,
	tm common.Transaction,
	pub domain.SymbolEventPublisher,
	logger log.Logger,
) domain.SymbolUseCase {
	return &useCase{
		repo:      repo,
		pub:       pub,
		validator: v,
		tm:        tm,
		log:       log.NewHelper(logger),
	}
}
```

**Publish events within transactions**:
```go
err := uc.tm.InTx(ctx, func(ctx context.Context) error {
	symbol, err = uc.repo.Create(ctx, s)
	if err != nil {
		return err
	}
	return uc.pub.PublishSymbolCreated(ctx, symbol)
})
```

**For delete operations, fetch entity first**:
```go
err := uc.tm.InTx(ctx, func(ctx context.Context) error {
	// Fetch entity before deletion (for event payload)
	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	err = uc.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return uc.pub.PublishSymbolDeleted(ctx, symbol)
})
```

**Why within transactions?**
- Ensures data consistency
- If publishing fails, database changes roll back
- Prevents "phantom events" for operations that didn't complete
</event_publishing_pattern>

<godoc_pattern>
## Documentation Pattern

**Constructor**:
```go
// New{Entity}UseCase creates a new {Entity}UseCase instance.
func New{Entity}UseCase(...) {Entity}UseCase
```

**Methods** (match interface comments):
```go
// Create{Entity} creates a new {Entity}.
func (uc *{entity}UseCase) Create{Entity}(ctx context.Context, e *{Entity}) (*{Entity}, error)
```

**Struct** (optional but recommended):
```go
// {entity}UseCase implements {Entity}UseCase interface.
type {entity}UseCase struct {
	...
}
```
</godoc_pattern>

<complete_example>
## Complete Use Case Example

```go
// Package symbol provides use cases for managing symbols.
package symbol

import (
	"context"
	"errors"
	"fmt"
	"platform/pagination"
	"symbols/internal/biz/domain"
	"symbols/internal/data/common"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
)

// useCase is a Symbol use case implementation.
type useCase struct {
	repo      domain.SymbolRepo
	pub       domain.SymbolEventPublisher
	log       *log.Helper
	validator *validator.Validate
	tm        common.Transaction
}

// NewUseCase creates a new Symbol use case.
func NewUseCase(
	repo domain.SymbolRepo,
	v *validator.Validate,
	tm common.Transaction,
	pub domain.SymbolEventPublisher,
	logger log.Logger,
) domain.SymbolUseCase {
	return &useCase{
		repo:      repo,
		validator: v,
		pub:       pub,
		tm:        tm,
		log:       log.NewHelper(logger),
	}
}

// GetSymbol gets a Symbol by its ID.
func (uc *useCase) GetSymbol(ctx context.Context, id uint64) (*domain.Symbol, error) {
	// Validate ID before calling repository
	if id <= 0 {
		return nil, domain.ErrInvalidID
	}

	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get symbol: %v", err)
		return nil, toDomainError(err)
	}

	return symbol, nil
}

// CreateSymbol creates a Symbol and returns the new Symbol.
func (uc *useCase) CreateSymbol(ctx context.Context, s *domain.Symbol) (*domain.Symbol, error) {
	// Validate domain model
	if err := uc.validator.Struct(s); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	var symbol *domain.Symbol

	err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		var err error
		symbol, err = uc.repo.Create(ctx, s)
		if err != nil {
			return err
		}

		return uc.pub.PublishSymbolCreated(ctx, symbol)
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Create error: %v", err)
		return nil, toDomainError(err)
	}

	return symbol, nil
}

// DeleteSymbol deletes a Symbol by its ID.
func (uc *useCase) DeleteSymbol(ctx context.Context, id uint64) error {
	// Validate ID
	if id <= 0 {
		return domain.ErrInvalidID
	}

	// Fetch symbol first (needed for event)
	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to find symbol: %v", err)
		return toDomainError(err)
	}

	// Delete with event publishing
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		err := uc.repo.Delete(ctx, id)
		if err != nil {
			return err
		}

		return uc.pub.PublishSymbolDeleted(ctx, symbol)
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete symbol: %v", err)
		return toDomainError(err)
	}

	return nil
}

// toDomainError transforms data layer errors to domain errors.
func toDomainError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, domain.ErrDataNotFound) {
		return domain.ErrSymbolNotFound
	}
	if errors.Is(err, domain.ErrDataDuplicateEntry) {
		return domain.ErrDuplicateSymbol
	}
	if errors.Is(err, domain.ErrDataDatabase) {
		return domain.ErrDatabaseOperation
	}

	return err
}
```
</complete_example>