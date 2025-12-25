---
name: kratos-repo
description: Implements go-kratos data layer repositories following Clean Architecture patterns with GORM, transactions, pagination, and error handling. Use when adding new data access layers to kratos microservices that need database persistence.
---

<objective>
Generate a complete, production-ready repository implementation for a go-kratos microservice that follows Clean Architecture principles with proper transaction handling, error mapping, pagination support, and data transformation.
</objective>

<context>
Repositories in go-kratos implement the data layer following Clean Architecture: they live in `internal/data/repo/`, implement interfaces from `internal/biz/interfaces.go`, and transform between data models (`internal/data/model/`) and domain models (`internal/biz/models.go`) while handling transactions and error mapping.
</context>

<quick_start>
To generate a repository for an entity named "Product":

1. Confirm `ProductRepo` interface exists in `internal/biz/interfaces.go`
2. Confirm `model.Product` exists in `internal/data/model/`
3. Confirm `biz.Product` domain model exists in `internal/biz/models.go`
4. Request: "Create repository for Product entity with Create, Update, FindByID, Delete, and ListProducts with pagination"
5. Skill generates `internal/data/repo/product.go` with all CRUD operations
6. Add constructor to `data.ProviderSet` in `internal/data/data.go`
7. Run `make generate` to wire dependencies
</quick_start>

<validation>
Before generating repository, verify:

- Entity interface exists in `internal/biz/interfaces.go` (e.g., `ProductRepo`)
- Domain model exists in `internal/biz/models.go` (e.g., `Product`)
- Entity model exists in `internal/data/model/` (e.g., `model.Product`)
- Service imports match pattern: `<service-name>/internal/...`
- Required packages are available: `gorm.io/gorm`, `github.com/go-kratos/kratos/v2/log`
- If pagination needed: `internal/pkg/pagination` package exists
</validation>

<process>
<step name="gather_requirements">
Ask the user for the entity name (e.g., "Symbol", "User", "Product"). This will be used throughout the implementation.

Confirm these assumptions:
- Repository will implement an interface from `internal/biz/interfaces.go`
- Entity model exists in `internal/data/model/`
- Domain model exists in `internal/biz/models.go`
- Common CRUD operations needed: Create, Update, FindByID, Delete
- List operation with pagination needed (yes/no)
</step>

<step name="create_repository_file">
Create `internal/data/repo/<entity_lowercase>.go` with package declaration and imports:

```go
package repo

import (
    "context"
    "errors"
    "strings"
    "<service-name>/internal/biz"
    "<service-name>/internal/data/common"
    "<service-name>/internal/data/model"
    "<service-name>/internal/pkg/pagination"  // Only if pagination needed

    "github.com/go-kratos/kratos/v2/log"
    "gorm.io/gorm"
)
```
</step>

<step name="generate_constructor_and_struct">
Create the constructor and struct following naming conventions:

```go
// New<Entity>Repo creates a new repository instance.
func New<Entity>Repo(db *gorm.DB, tx common.Transaction, logger log.Logger) biz.<Entity>Repo {
    return &<entity>Repo{
        db:  db,
        tx:  tx,
        log: log.NewHelper(logger),
    }
}

type <entity>Repo struct {
    db  *gorm.DB
    tx  common.Transaction
    log *log.Helper
}
```

**Naming rules:**
- Constructor: `New<Entity>Repo` (PascalCase, exported)
- Struct: `<entity>Repo` (camelCase, unexported)
- Fields: `db`, `tx`, `log` (exactly these names)
- Receiver: `r` (single letter)
</step>

<step name="implement_crud_operations">
Implement the requested CRUD operations. For each operation:

**Create:** Use transaction with `FullSaveAssociations`, rollback on error
**Update:** Use transaction, check `RowsAffected == 0`, fetch and return updated entity
**FindByID:** No transaction, use `Preload()` for relationships
**Delete:** Use transaction, check `RowsAffected == 0`
**List:** No transaction, separate count and data queries, calculate pagination metadata

All write operations (Create/Update/Delete) must:
- Begin explicit transaction with `r.db.WithContext(ctx).Begin()`
- Defer panic recovery with rollback
- Rollback on any error
- Commit at the end

All read operations (FindByID/List) must:
- Use `r.db.WithContext(ctx)` directly (no transaction)
- Use `Preload()` for eager loading relationships

See `references/crud-templates.md` for complete code templates for each operation.
</step>

<step name="add_error_mapping">
Add error mapping helpers:

```go
func (r *<entity>Repo) mapGormError(err error) error {
    if err == nil {
        return nil
    }
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return biz.ErrNotFound
    }
    if r.isDuplicateKeyError(err) {
        return biz.ErrDuplicateEntry
    }
    return biz.ErrDatabase
}

func (r *<entity>Repo) isDuplicateKeyError(err error) bool {
    errMsg := err.Error()
    return strings.Contains(errMsg, "Error 1062") ||
        strings.Contains(errMsg, "Duplicate entry")
}
```

**Error mapping:**
- `gorm.ErrRecordNotFound` → `biz.ErrNotFound`
- MySQL "Error 1062" or "Duplicate entry" → `biz.ErrDuplicateEntry`
- All other errors → `biz.ErrDatabase`

See `references/error-handling.md` for complete error handling patterns.
</step>

<step name="add_data_transformers">
Add data transformation functions:

```go
func toDomain<Entity>(e *model.<Entity>) *biz.<Entity> {
    if e == nil {
        return nil
    }
    // Map all fields from model to domain
    // Handle nested relationships with nil checks
    return &biz.<Entity>{...}
}

func toEntity<Entity>(d *biz.<Entity>) *model.<Entity> {
    if d == nil {
        return nil
    }
    // Map all fields from domain to model
    // Handle nested relationships with nil checks
    return &model.<Entity>{...}
}
```

**Transformation rules:**
- Function names: `toDomain<Entity>` and `toEntity<Entity>` (unexported)
- Always check for nil at the start
- Map ID fields: `model.ID` ↔ `biz.Id` (note capitalization difference)
- Handle nested relationships with nil checks
</step>

<step name="update_provider_set">
Add the repository constructor to `internal/data/data.go`:

```go
var ProviderSet = wire.NewSet(
    NewDB,
    NewData,
    NewTransaction,
    repo.New<Entity>Repo,  // Add this line
)
```
</step>

<step name="verify_wire_integration">
Run `make generate` to regenerate wire dependency injection. The constructor will be automatically wired with:
- `db *gorm.DB` from `NewDB`
- `tx common.Transaction` from `NewTransaction`
- `logger log.Logger` from app initialization
</step>
</process>

<error_handling>
All repository methods must follow these error handling patterns:

<pattern name="error_logging">
Every error must be logged before returning:

```go
if err != nil {
    r.log.WithContext(ctx).Errorf("Failed to <operation> <entity>: %v", err)
    return nil, r.mapGormError(err)
}
```
</pattern>

<pattern name="error_mapping">
All GORM errors must be mapped through `mapGormError()`:
- `gorm.ErrRecordNotFound` → `biz.ErrNotFound`
- MySQL duplicate key (Error 1062) → `biz.ErrDuplicateEntry`
- All other errors → `biz.ErrDatabase`
</pattern>

<pattern name="rows_affected_check">
Update and Delete operations must check for zero rows:

```go
if result.RowsAffected == 0 {
    r.log.WithContext(ctx).Errorf("Failed to <operation> <entity>: 0 rows affected")
    tx.Rollback()
    return biz.ErrNotFound
}
```
</pattern>

<pattern name="transaction_rollback">
All write operations must rollback on error:

```go
if err := tx.Create(entity).Error; err != nil {
    tx.Rollback()
    r.log.WithContext(ctx).Errorf("Failed to save <entity>: %v", err)
    return nil, r.mapGormError(err)
}
```
</pattern>

For detailed error handling patterns, see `references/error-handling.md`.
</error_handling>

<references>
For detailed guidance on specific topics, see:

- **references/naming-conventions.md**: Strict naming rules for structs, fields, methods, and helpers
- **references/transaction-patterns.md**: Transaction handling for write vs read operations
- **references/error-handling.md**: Error mapping, logging, and RowsAffected checks
- **references/crud-templates.md**: Complete code templates for all CRUD operations
- **references/complete-examples.md**: Full working repository implementations
</references>

<success_criteria>
Repository implementation is complete when:

- File created at `internal/data/repo/<entity_lowercase>.go`
- Package declaration is `package repo`
- Constructor follows pattern: `func New<Entity>Repo(db *gorm.DB, tx common.Transaction, logger log.Logger) biz.<Entity>Repo`
- Struct has exactly three fields: `db`, `tx`, `log`
- All write operations (Create/Update/Delete) use explicit transactions
- All read operations use `r.db.WithContext(ctx)` without transactions
- Update and Delete check `RowsAffected == 0` and return `biz.ErrNotFound`
- All errors are logged with context before returning
- All GORM errors are mapped through `mapGormError()`
- Pagination calculates `HasNextPage` and `HasPreviousPage` correctly
- Data transformers handle nil checks
- Repository constructor added to `data.ProviderSet` in `internal/data/data.go`
- `make generate` runs successfully and generates wire bindings
- Code follows exact naming conventions (case-sensitive)
</success_criteria>