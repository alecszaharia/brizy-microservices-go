package biz

import (
	"context"
	"platform/pagination"
)

type SymbolUseCase interface {
	// GetSymbol retrieves a Symbol by its ID.
	GetSymbol(ctx context.Context, id uint64) (*Symbol, error)

	// CreateSymbol creates a new Symbol.
	CreateSymbol(ctx context.Context, g *Symbol) (*Symbol, error)

	// UpdateSymbol updates an existing Symbol and returns the updated Symbol.
	UpdateSymbol(ctx context.Context, g *Symbol) (*Symbol, error)

	// DeleteSymbol deletes a Symbol by its ID.
	DeleteSymbol(ctx context.Context, id uint64) error

	// ListSymbols lists Symbols based on the provided options and returns pagination metadata.
	ListSymbols(ctx context.Context, params *pagination.OffsetPaginationParams, filter map[string]interface{}) ([]*Symbol, *pagination.Meta, error)
}

type SymbolRepo interface {

	// Create saves the given Symbol.
	Create(context.Context, *Symbol) (*Symbol, error)

	// Update updates a Symbol in the repository.
	Update(context.Context, *Symbol) (*Symbol, error)

	// FindByID returns the Symbol with the given ID from the repository.
	FindByID(context.Context, uint64) (*Symbol, error)

	// ListSymbols returns a list of Symbols from the repository with pagination metadata.
	ListSymbols(context.Context, uint64, uint32, map[string]interface{}) ([]*Symbol, *pagination.Meta, error)

	// Delete removes a Symbol from the repository by its ID. Returns an error if the operation fails.
	Delete(context.Context, uint64) error
}
