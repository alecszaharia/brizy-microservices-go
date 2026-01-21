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
func NewUseCase(repo domain.SymbolRepo, v *validator.Validate, tm common.Transaction, pub domain.SymbolEventPublisher, logger log.Logger) domain.SymbolUseCase {
	return &useCase{repo: repo, validator: v, pub: pub, tm: tm, log: log.NewHelper(logger)}
}

// GetSymbol gets a Symbol by its ID.
func (uc *useCase) GetSymbol(ctx context.Context, id uint64) (*domain.Symbol, error) {

	// Validate ID before calling a repository
	if id <= 0 {
		return nil, domain.ErrInvalidID
	}

	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		// Transform data layer error to domain error
		uc.log.WithContext(ctx).Errorf("Failed to get symbol: %v", err)

		return nil, toDomainError(err)
	}

	return symbol, nil
}

// CreateSymbol creates a Symbol and returns the new Symbol.
func (uc *useCase) CreateSymbol(ctx context.Context, g *domain.Symbol) (*domain.Symbol, error) {
	// Validate domain model
	if err := uc.validator.Struct(g); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	var symbol *domain.Symbol

	err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		var err error
		symbol, err = uc.repo.Create(ctx, g)

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

// UpdateSymbol updates an existing Symbol and returns the updated Symbol.
func (uc *useCase) UpdateSymbol(ctx context.Context, g *domain.Symbol) (*domain.Symbol, error) {
	// Validate ID
	if g.ID <= 0 {
		return nil, domain.ErrInvalidID
	}

	// Validate domain model
	if err := uc.validator.Struct(g); err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}
	var updatedSymbol *domain.Symbol
	// Proceed to update
	err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		var err error
		updatedSymbol, err = uc.repo.Update(ctx, g)

		if err != nil {
			return err
		}

		return uc.pub.PublishSymbolUpdated(ctx, updatedSymbol)
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to update symbol: %v", err)

		return nil, toDomainError(err)
	}

	return updatedSymbol, nil
}

// DeleteSymbol deletes a Symbol by its ID.
func (uc *useCase) DeleteSymbol(ctx context.Context, id uint64) error {
	// Validate ID
	if id <= 0 {
		return domain.ErrInvalidID
	}

	symbol, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to find symbol: %v", err)
		return toDomainError(err)
	}

	// Proceed to delete
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {

		err := uc.repo.Delete(ctx, id)

		if err != nil {
			return err
		}

		return uc.pub.PublishSymbolDeleted(ctx, symbol)
	})

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete the symbol: %v", err)

		return toDomainError(err)
	}

	return nil
}

// ListSymbols lists Symbols based on the provided options with pagination metadata.
func (uc *useCase) ListSymbols(ctx context.Context, params *pagination.OffsetPaginationParams, filter map[string]interface{}) ([]*domain.Symbol, *pagination.Meta, error) {

	// Validate options (including params )
	if err := uc.validator.Struct(params); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", domain.ErrValidationFailed, err)
	}

	// Call a repository to get symbols and params metadata
	symbols, meta, err := uc.repo.ListSymbols(ctx, params.Offset, params.Limit, filter)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to list symbols: %v", err)
		return nil, nil, toDomainError(err)
	}

	return symbols, meta, nil
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
	if errors.Is(err, domain.ErrDataTransactionFailed) {
		return domain.ErrDatabaseOperation
	}
	if errors.Is(err, domain.ErrDataDatabase) {
		return domain.ErrDatabaseOperation
	}

	return err
}
