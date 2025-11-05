package data

import (
	"context"
	"fmt"
	"time"

	"symbol-service/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type SymbolEntity struct {
	ID              int32             `gorm:"primaryKey;autoIncrement;index:idx_project_id,priority:2" json:"id"`
	ProjectID       int32             `gorm:"not null;uniqueIndex:idx_project_uid,priority:1;index:idx_project_id,priority:1" json:"project_id"`
	UID             string            `gorm:"not null;size:255;uniqueIndex:idx_project_uid,priority:2" json:"uid"`
	Label           string            `gorm:"not null;size:255" json:"label"`
	ClassName       string            `gorm:"not null;size:255" json:"class_name"`
	ComponentTarget string            `gorm:"not null;size:255" json:"component_target"`
	Version         string            `gorm:"not null;size:50" json:"version"`
	SymbolData      *SymbolDataEntity `gorm:"foreignKey:SymbolID;references:ID;constraint:OnDelete:CASCADE" json:"symbol_data,omitempty"`
	CreatedAt       time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

func (SymbolEntity) TableName() string {
	return "symbols"
}

type SymbolDataEntity struct {
	ID        int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	SymbolID  int32     `gorm:"not null;uniqueIndex" json:"symbol_id"`
	Data      *[]byte   `gorm:"not null;type:longblob" json:"data"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SymbolDataEntity) TableName() string {
	return "symbol_data"
}

type symbolRepo struct {
	data *Data
	log  *log.Helper
}

// NewSymbolRepo .
func NewSymbolRepo(data *Data, logger log.Logger) biz.SymbolRepo {
	return &symbolRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *symbolRepo) Save(ctx context.Context, g *biz.Symbol) (*biz.Symbol, error) {
	symbol := toEntitySymbol(g)

	// Begin transaction
	tx := r.data.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create symbol
	if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(symbol).Error; err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to save symbol: %v", err)
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to commit transaction: %v", err)
		return nil, err
	}

	return toDomainSymbol(symbol), nil
}

func (r *symbolRepo) Update(ctx context.Context, g *biz.Symbol) (*biz.Symbol, error) {
	symbol := toEntitySymbol(g)

	// Begin transaction
	tx := r.data.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update symbol
	result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(&SymbolEntity{}).Where("id = ?", symbol.ID).Updates(symbol)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.NotFound("SYMBOL_NOT_FOUND", "symbol not found")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to commit transaction: %v", err)
		return nil, err
	}

	// Fetch updated symbol with data
	return r.FindByID(ctx, symbol.ID)
}

func (r *symbolRepo) FindByID(ctx context.Context, id int32) (*biz.Symbol, error) {
	var symbol SymbolEntity

	err := r.data.db.WithContext(ctx).
		Preload("SymbolData").
		Where("id = ?", id).
		First(&symbol).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("SYMBOL_NOT_FOUND", "symbol not found") // FIXED: return proper error
		}
		r.log.WithContext(ctx).Errorf("Failed to find symbol by ID %d: %v", id, err)
		return nil, err
	}

	return toDomainSymbol(&symbol), nil
}

// ListSymbols returns symbols WITHOUT symbol data for performance reasons.
// Use FindByID to retrieve individual symbols with their data.
func (r *symbolRepo) ListSymbols(ctx context.Context, options *biz.ListSymbolsOptions) ([]*biz.Symbol, *biz.SymbolCursor, error) {
	var symbolEntites []SymbolEntity

	query := r.data.db.WithContext(ctx).Where("project_id = ?", options.ProjectID)

	// filter the entities by cursor
	if options.Cursor != nil {
		switch options.Cursor.OrderBy {
		case "id":
			if options.Cursor.Direction == "ASC" {
				query = query.Where("id > ?", options.Cursor.LastID)
			} else {
				query = query.Where("id < ?", options.Cursor.LastID)
			}
		case "label":
			if options.Cursor.Direction == "ASC" {
				query = query.Where("label > ? OR (label = ? AND id > ?)",
					options.Cursor.LastValue, options.Cursor.LastValue, options.Cursor.LastID)
			} else {
				query = query.Where("label < ? OR (label = ? AND id < ?)",
					options.Cursor.LastValue, options.Cursor.LastValue, options.Cursor.LastID)
			}
		case "created_at":
			if options.Cursor.Direction == "ASC" {
				query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
					options.Cursor.LastValue, options.Cursor.LastValue, options.Cursor.LastID)
			} else {
				query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
					options.Cursor.LastValue, options.Cursor.LastValue, options.Cursor.LastID)
			}
		}
		query = query.Order(fmt.Sprintf("%s %s, id %s",
			options.Cursor.OrderBy, options.Cursor.Direction, options.Cursor.Direction))

	} else {
		// Default ordering when no cursor
		query = query.Order("id ASC")
	}

	// Apply limit (fetch PageSize + 1 to detect if there are more results)
	query = query.Limit(int(options.PageSize) + 1)

	// Execute the query
	err := query.Find(&symbolEntites).Error
	if err != nil {
		r.log.WithContext(ctx).Errorf("Failed to list symbols: %v", err)
		return nil, nil, err
	}

	// Build result and cursor
	var result []*biz.Symbol
	var nextCursor *biz.SymbolCursor

	hasMore := len(symbolEntites) > int(options.PageSize)
	if hasMore {
		symbolEntites = symbolEntites[:options.PageSize]
	}

	for _, symbol := range symbolEntites {
		result = append(result, toDomainSymbol(&symbol))
	}

	// Build next cursor if there are more results
	if hasMore && len(symbolEntites) > 0 {
		last := symbolEntites[len(symbolEntites)-1]
		orderBy := "id"
		direction := "ASC"
		lastValue := ""

		if options.Cursor != nil {
			orderBy = options.Cursor.OrderBy
			direction = options.Cursor.Direction

			switch orderBy {
			case "label":
				lastValue = last.Label
			case "created_at":
				lastValue = last.CreatedAt.Format(time.RFC3339)
			}
		}

		nextCursor = &biz.SymbolCursor{
			LastID:    int64(last.ID),
			LastValue: lastValue,
			OrderBy:   orderBy,
			Direction: direction,
		}
	}

	return result, nextCursor, nil
}

func (r *symbolRepo) Delete(ctx context.Context, id int32) error {
	// Begin transaction
	tx := r.data.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete symbol
	result := tx.Where("id = ?", id).Delete(&SymbolEntity{})
	if err := result.Error; err != nil {
		tx.Rollback()
		r.log.WithContext(ctx).Errorf("Failed to delete symbol: %v", err)
		return err
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.NotFound("SYMBOL_NOT_FOUND", "symbol not found")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		r.log.WithContext(ctx).Errorf("Failed to commit delete transaction: %v", err)
		return err
	}

	return nil
}

// toDomainSymbol converts data.SymbolEntity to biz.Symbol
func toDomainSymbol(s *SymbolEntity) *biz.Symbol {
	if s == nil {
		return nil
	}

	domain := &biz.Symbol{
		Id:              s.ID,
		Project:         s.ProjectID,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
	}

	if s.SymbolData != nil {
		domain.Data = &biz.SymbolData{
			Id:      s.SymbolData.ID,
			Project: s.ProjectID,
			Data:    s.SymbolData.Data,
		}
	}

	return domain
}

// toEntitySymbol converts biz.Symbol to data.SymbolEntity
func toEntitySymbol(s *biz.Symbol) *SymbolEntity {
	if s == nil {
		return nil
	}

	data := &SymbolEntity{
		ID:              s.Id,
		ProjectID:       s.Project,
		UID:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
	}

	if s.Data != nil {
		data.SymbolData = &SymbolDataEntity{
			ID:       s.Data.Id,
			SymbolID: s.Id,
			Data:     s.Data.Data,
		}
	}

	return data
}
