# Interface Pattern

**Location**: All interfaces are defined in `internal/biz/domain/interfaces.go`

<use_case_interface>
## UseCase Interface

**Purpose**: Defines business operations exposed to service layer.

**Pattern**:
```go
type {Entity}UseCase interface {
	// Get{Entity} retrieves a {Entity} by its ID.
	Get{Entity}(ctx context.Context, id uint64) (*{Entity}, error)

	// Create{Entity} creates a new {Entity}.
	Create{Entity}(ctx context.Context, e *{Entity}) (*{Entity}, error)

	// Update{Entity} updates an existing {Entity}.
	Update{Entity}(ctx context.Context, e *{Entity}) (*{Entity}, error)

	// Delete{Entity} deletes a {Entity} by its ID.
	Delete{Entity}(ctx context.Context, id uint64) error

	// List{Entities} lists {Entities} with pagination.
	List{Entities}(ctx context.Context, options *List{Entities}Options) ([]*{Entity}, *pagination.PaginationMeta, error)
}
```

**Rules**:
- One interface per entity
- Methods match use case implementation
- First parameter always `context.Context`
- Return errors as last return value
- Use pointer types for structs
</use_case_interface>

<repo_interface>
## Repository Interface

**Purpose**: Defines data access contract implemented by data layer.

**Pattern**:
```go
type {Entity}Repo interface {
	// Create saves the given {Entity}.
	Create(context.Context, *{Entity}) (*{Entity}, error)

	// Update updates a {Entity}.
	Update(context.Context, *{Entity}) (*{Entity}, error)

	// FindByID returns the {Entity} with the given ID.
	FindByID(context.Context, uint64) (*{Entity}, error)

	// List{Entities} returns a list of {Entities}.
	List{Entities}(context.Context, *List{Entities}Options) ([]*{Entity}, *pagination.PaginationMeta, error)

	// Delete removes a {Entity} by its ID.
	Delete(context.Context, uint64) error
}
```

**Rules**:
- Repo methods use short names (Create vs Create{Entity})
- Accept/return business models, not ORM entities
- First parameter always `context.Context`
- No business logic in repo interface
</repo_interface>

<event_publisher_interface>
## Event Publisher Interface (Optional)

**Purpose**: Defines event publishing contract for domain events.

**Pattern**:
```go
type {Entity}EventPublisher interface {
	// PublishSymbolCreated publishes a SymbolCreated event.
	PublishSymbolCreated(ctx context.Context, symbol *Symbol) error

	// PublishSymbolUpdated publishes a SymbolUpdated event.
	PublishSymbolUpdated(ctx context.Context, symbol *Symbol) error

	// PublishSymbolDeleted publishes a SymbolDeleted event.
	PublishSymbolDeleted(ctx context.Context, symbol *Symbol) error
}
```

**Rules**:
- Methods accept domain models (not protobuf or events)
- First parameter always `context.Context`
- Return error for publish failures
- Naming: `Publish{Entity}{Action}` (e.g., PublishSymbolCreated)
- Accept full entity (not just ID) for rich event payload

**When to use**:
- Services with event-driven architecture
- When other services need to react to domain events
- For audit logging and event sourcing patterns
</event_publisher_interface>

<naming_rules>
## Naming Rules

**Interface Names** (all in `domain` package):
- UseCase Interface: `{Entity}UseCase` (e.g., `SymbolUseCase`)
- Repo Interface: `{Entity}Repo` (e.g., `SymbolRepo`)
- Event Publisher: `{Entity}EventPublisher` (e.g., `SymbolEventPublisher`)

**Method Naming**:
- UseCase: `Create{Entity}`, `Get{Entity}`, `Update{Entity}`, `Delete{Entity}`, `List{Entities}`
- Repo: `Create`, `FindByID`, `Update`, `Delete`, `List{Entities}`
- Event Publisher: `Publish{Entity}{Action}` (e.g., `PublishSymbolCreated`)

**Why Different?**:
- UseCase methods are entity-specific (exported to service layer)
- Repo methods are generic (internal to biz layer)
- Event publisher methods are explicit about event type

**Package Structure**:
- Interfaces defined in: `internal/biz/domain/interfaces.go`
- Implementation packages: `internal/biz/{entity}/usecase.go`
- Import as: `import "symbols/internal/biz/domain"`
</naming_rules>
