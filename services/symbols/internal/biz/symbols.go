package biz

import (
	"context"
	"errors"
	"fmt"
	"platform/events"
	"platform/pagination"
	"symbols/internal/data/common"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// SymbolUseCase is a Symbol use case.
type symbolUseCase struct {
	repo      SymbolRepo
	pub       events.Publisher
	log       *log.Helper
	validator *validator.Validate
	tm        common.Transaction
}

// NewSymbolUseCase creates a new Symbol use case.
func NewSymbolUseCase(repo SymbolRepo, validator *validator.Validate, tm common.Transaction, pub events.Publisher, logger log.Logger) SymbolUseCase {
	return &symbolUseCase{repo: repo, validator: validator, pub: pub, tm: tm, log: log.NewHelper(logger)}
}

// GetSymbol gets a Symbol by its ID.
func (uc *symbolUseCase) GetSymbol(ctx context.Context, id uint64) (*Symbol, error) {

	// Validate ID before calling a repository
	if id <= 0 {
		return nil, ErrInvalidID
	}

	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		// Transform data layer error to domain error
		uc.log.WithContext(ctx).Errorf("Failed to get symbol: %v", err)

		if errors.Is(err, ErrNotFound) {
			return nil, ErrSymbolNotFound
		}
		// Any other database error
		return nil, ErrDatabaseOperation
	}

	return symbol, nil
}

// CreateSymbol creates a Symbol and returns the new Symbol.
func (uc *symbolUseCase) CreateSymbol(ctx context.Context, g *Symbol) (*Symbol, error) {
	// Validate domain model
	if err := uc.validator.Struct(g); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	var symbol *Symbol
	var err error

	err = uc.tm.InTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
		var err error
		symbol, err = uc.repo.Create(ctx, g)

		if err != nil {
			return err
		}

		err2 := uc.pub.Publish(ctx, "weweew", make([]byte, 10))
		if err2 != nil {
			return err2
		}

		return err
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Create error: %v", err)

		if errors.Is(err, ErrDuplicateEntry) {
			return nil, ErrDuplicateSymbol
		}

		// Generic database error
		return nil, ErrDatabaseOperation
	}

	return symbol, nil
}

// UpdateSymbol updates an existing Symbol and returns the updated Symbol.
func (uc *symbolUseCase) UpdateSymbol(ctx context.Context, g *Symbol) (*Symbol, error) {
	// Validate ID
	if g.Id <= 0 {
		return nil, ErrInvalidID
	}

	// Validate domain model
	if err := uc.validator.Struct(g); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Proceed to update
	symbol, err := uc.repo.Update(ctx, g)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to update symbol: %v", err)

		// Check if symbol was not found (RowsAffected == 0)
		if errors.Is(err, ErrNotFound) {
			return nil, ErrSymbolNotFound
		}

		// Check for duplicate entry error from data layer
		if errors.Is(err, ErrDuplicateEntry) {
			return nil, ErrDuplicateSymbol
		}

		// Generic database error
		return nil, ErrDatabaseOperation
	}

	return symbol, nil
}

// DeleteSymbol deletes a Symbol by its ID.
func (uc *symbolUseCase) DeleteSymbol(ctx context.Context, id uint64) error {
	// Validate ID
	if id <= 0 {
		return ErrInvalidID
	}

	err := uc.repo.Delete(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete the symbol: %v", err)

		// Check if symbol was not found (RowsAffected == 0)
		if errors.Is(err, ErrNotFound) {
			return ErrSymbolNotFound
		}

		// Generic database error
		return ErrDatabaseOperation
	}

	return nil
}

// ListSymbols lists Symbols based on the provided options with pagination metadata.
func (uc *symbolUseCase) ListSymbols(ctx context.Context, params *pagination.OffsetPaginationParams, filter map[string]interface{}) ([]*Symbol, *pagination.Meta, error) {

	// Validate options (including params )
	if err := uc.validator.Struct(params); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Call a repository to get symbols and params metadata
	symbols, meta, err := uc.repo.ListSymbols(ctx, params.Offset, params.Limit, filter)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to list symbols: %v", err)
		return nil, nil, ErrDatabaseOperation
	}

	return symbols, meta, nil
}
