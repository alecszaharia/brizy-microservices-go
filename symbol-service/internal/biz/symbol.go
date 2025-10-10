package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
)

// Symbol is a Symbol model.
type Symbol struct {
	Id              int32       `validate:"omitempty,gte=0"`
	Project         int32       `validate:"required,gt=0"`
	Uid             string      `validate:"required,uuid4"`
	Label           string      `validate:"required,min=1,max=255"`
	ClassName       string      `validate:"required,min=1,max=255"`
	ComponentTarget string      `validate:"required,min=1,max=100"`
	Version         string      `validate:"required,semver"`
	Data            *SymbolData `validate:"required"`
}

type SymbolData struct {
	Id      int32   `validate:"omitempty,gte=0"`
	Project int32   `validate:"required,gt=0"`
	Data    *[]byte `validate:"omitempty"`
}

type SymbolCursor struct {
	LastID    int64  `json:"id" validate:"required,gte=0"`
	LastValue string `json:"value" validate:"omitempty,max=255"` // Or use a union type
	OrderBy   string `json:"order_by" validate:"required,oneof=id label created_at"`
	Direction string `json:"direction" validate:"required,oneof=ASC DESC"`
}

type ListSymbolsOptions struct {
	ProjectID int32         `validate:"required,gt=0"`
	PageSize  int32         `validate:"required,gte=1,lte=100"`
	Cursor    *SymbolCursor `validate:"omitempty"`
}

type SymbolRepo interface {
	Save(context.Context, *Symbol) (*Symbol, error)
	Update(context.Context, *Symbol) (*Symbol, error)
	FindByID(context.Context, int32) (*Symbol, error)
	ListSymbols(context.Context, *ListSymbolsOptions) ([]*Symbol, *SymbolCursor, error)
	Delete(context.Context, int32) error
}

// SymbolUseCase is a Symbol use case.
type SymbolUseCase struct {
	repo      SymbolRepo
	log       *log.Helper
	validator *validator.Validate
}

// NewSymbolUseCase new a Symbol usecase.
func NewSymbolUseCase(repo SymbolRepo, validator *validator.Validate, logger log.Logger) *SymbolUseCase {
	return &SymbolUseCase{repo: repo, validator: validator, log: log.NewHelper(logger)}
}

// CreateSymbol creates a Symbol, and returns the new Symbol.
func (uc *SymbolUseCase) CreateSymbol(ctx context.Context, g *Symbol) (*Symbol, error) {
	uc.log.WithContext(ctx).Infof("CreateSymbol: %v", g.Label)

	if err := uc.validator.Struct(g); err != nil {
		return nil, err
	}

	symbol, err := uc.repo.Save(ctx, g)

	if err != nil {
		uc.log.WithContext(ctx).Errorf("Create error: %v", err)
		return nil, err
	}

	return symbol, nil
}

// UpdateSymbol updates an existing Symbol and returns the updated Symbol.
func (uc *SymbolUseCase) UpdateSymbol(ctx context.Context, g *Symbol) (*Symbol, error) {
	uc.log.WithContext(ctx).Infof("UpdateSymbol: %v", g.Id)
	// Check if the symbol exists before updating

	if err := uc.validator.Struct(g); err != nil {
		return nil, err
	}

	// Proceed to update
	updatedSymbol, err := uc.repo.Update(ctx, g)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to update symbol: %v", err)
		return nil, err
	}
	return updatedSymbol, nil
}

// DeleteSymbol deletes a Symbol by its ID.
func (uc *SymbolUseCase) DeleteSymbol(ctx context.Context, id int32) error {

	if id <= 0 {
		return errors.BadRequest("INVALID_ID", "Symbol ID must be positive")
	}

	err := uc.repo.Delete(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete the symbol: %v", err)
		return err
	}
	return nil
}

// ListSymbols lists Symbols based on the provided options.
func (uc *SymbolUseCase) ListSymbols(ctx context.Context, options *ListSymbolsOptions) ([]*Symbol, *SymbolCursor, error) {

	if err := uc.validator.Struct(options); err != nil {
		return nil, nil, err
	}

	list, cursor, err := uc.repo.ListSymbols(ctx, options)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to obtain the symbol list")
		return nil, nil, err
	}
	return list, cursor, nil
}
