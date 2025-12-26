package repo

import (
	"brizy-go-platform/pagination"
	"context"
	"errors"
	"strings"
	"symbols/internal/biz"
	"symbols/internal/data/common"
	"symbols/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// NewSymbolRepo .
func NewSymbolRepo(db *gorm.DB, tx common.Transaction, logger log.Logger) biz.SymbolRepo {
	return &symbolRepo{
		db:  db,
		tx:  tx,
		log: log.NewHelper(logger),
	}
}

type symbolRepo struct {
	db  *gorm.DB
	tx  common.Transaction
	log *log.Helper
}

func (r *symbolRepo) Create(ctx context.Context, s *biz.Symbol) (*biz.Symbol, error) {

	symbol := toEntitySymbol(s)

	// Create symbol
	if err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(symbol).Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to save symbol: %v", err)
		return nil, r.mapGormError(err)
	}

	return toDomainSymbol(symbol), nil
}

func (r *symbolRepo) Update(ctx context.Context, symbol *biz.Symbol) (*biz.Symbol, error) {

	// Transform to entity
	entity := toEntitySymbol(symbol)

	// Update symbol
	result := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Model(&model.Symbol{}).Where("id = ?", symbol.Id).Updates(entity)
	if result.Error != nil {
		return nil, r.mapGormError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, biz.ErrNotFound
	}

	// Fetch updated symbol with data
	return r.FindByID(ctx, symbol.Id)
}

func (r *symbolRepo) FindByID(ctx context.Context, id uint64) (*biz.Symbol, error) {
	var symbol *model.Symbol

	err := r.db.WithContext(ctx).
		Preload("SymbolData").
		Where("id = ?", id).
		First(&symbol).Error

	if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to find symbol by ID %d: %v", id, err)
		return nil, r.mapGormError(err)
	}

	return toDomainSymbol(symbol), nil
}

func (r *symbolRepo) ListSymbols(ctx context.Context, options *biz.ListSymbolsOptions) ([]*biz.Symbol, *pagination.PaginationMeta, error) {
	var symbolEntities []*model.Symbol
	var totalCount int64

	// Base query with project filter
	baseQuery := r.db.WithContext(ctx).Model(&model.Symbol{}).Where("project_id = ?", options.ProjectID)

	// Execute count query for total_count
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to count symbols: %v", err)
		return nil, nil, r.mapGormError(err)
	}

	// Execute data query with pagination
	query := baseQuery.
		Limit(int(options.Pagination.Limit)).  // Limit results
		Offset(int(options.Pagination.Offset)) // Skip offset records
	//Preload("SymbolData")                   // Eagerly load symbol data - I think this should not be loaded for list

	if err := query.Find(&symbolEntities).Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to list symbols: %v", err)
		return nil, nil, r.mapGormError(err)
	}

	// Transform entities to domain objects
	symbols := make([]*biz.Symbol, 0, len(symbolEntities))
	for _, entity := range symbolEntities {
		symbols = append(symbols, toDomainSymbol(entity))
	}

	// Calculate pagination metadata
	meta := &pagination.PaginationMeta{
		TotalCount:      uint64(totalCount),
		Offset:          options.Pagination.Offset,
		Limit:           options.Pagination.Limit,
		HasNextPage:     options.Pagination.Offset+uint64(len(symbolEntities)) < uint64(totalCount),
		HasPreviousPage: options.Pagination.Offset > 0,
	}

	return symbols, meta, nil
}

func (r *symbolRepo) Delete(ctx context.Context, id uint64) error {
	// Begin transaction

	// Delete symbol
	result := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", id).Delete(&model.Symbol{})
	if err := result.Error; err != nil {
		return r.mapGormError(err)
	}

	if result.RowsAffected == 0 {
		return biz.ErrNotFound
	}

	return nil
}

// Internal helper functions

// mapGormError translates GORM errors to data layer errors
func (r *symbolRepo) mapGormError(err error) error {
	if err == nil {
		return nil
	}

	// Check for GORM-specific errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return biz.ErrNotFound
	}

	// Check for duplicate key constraint violations
	if r.isDuplicateKeyError(err) {
		return biz.ErrDuplicateEntry
	}

	// Wrap other database errors
	return biz.ErrDatabase
}

// isDuplicateKeyError checks if error is a duplicate key violation
func (r *symbolRepo) isDuplicateKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "Error 1062") ||
		strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "UNIQUE constraint failed")
}

// toDomainSymbol converts a model.Symbol entity to its corresponding biz.Symbol domain object.
func toDomainSymbol(s *model.Symbol) *biz.Symbol {
	if s == nil {
		return nil
	}

	d := &biz.Symbol{
		Id:              s.ID,
		Project:         s.ProjectID,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
	}

	if s.SymbolData != nil {
		d.Data = &biz.SymbolData{
			Id:      s.SymbolData.ID,
			Project: s.ProjectID,
			Data:    s.SymbolData.Data,
		}
	}

	return d
}

// toEntitySymbol transforms a *biz.Symbol domain object into a *model.Symbol persistence object.
// Returns nil if the input is nil.
func toEntitySymbol(s *biz.Symbol) *model.Symbol {
	if s == nil {
		return nil
	}

	d := &model.Symbol{
		BaseModel: model.BaseModel{
			ID: s.Id,
		},
		ProjectID:       s.Project,
		UID:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
	}

	if s.Data != nil {
		d.SymbolData = &model.SymbolData{
			BaseModel: model.BaseModel{
				ID: s.Data.Id,
			},
			SymbolID: s.Id,
			Data:     s.Data.Data,
		}
	}

	return d
}
