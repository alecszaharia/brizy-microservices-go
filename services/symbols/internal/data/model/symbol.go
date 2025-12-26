package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index;index:idx_project_deleted_at,priority:2"`
}

type Symbol struct {
	BaseModel
	ProjectID       uint64      `gorm:"not null;uniqueIndex:idx_project_uid,priority:1;index:idx_project_id,priority:1;index:idx_project_deleted_at,priority:1" json:"project_id"`
	UID             string      `gorm:"not null;size:255;uniqueIndex:idx_project_uid,priority:2" json:"uid"`
	Label           string      `gorm:"not null;size:255" json:"label"`
	ClassName       string      `gorm:"not null;size:255" json:"class_name"`
	ComponentTarget string      `gorm:"not null;size:255" json:"component_target"`
	Version         uint32      `gorm:"not null" json:"version"`
	SymbolData      *SymbolData `gorm:"foreignKey:SymbolID;references:ID;constraint:OnDelete:CASCADE" json:"symbol_data,omitempty"`
}

func (Symbol) TableName() string {
	return "symbols"
}

type SymbolData struct {
	BaseModel
	SymbolID uint64  `gorm:"not null;uniqueIndex" json:"symbol_id"`
	Data     *[]byte `gorm:"not null;type:longblob" json:"data"`
}

func (SymbolData) TableName() string {
	return "symbol_data"
}
