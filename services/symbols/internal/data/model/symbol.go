package model

import (
	"time"

	"gorm.io/gorm"
)

type Symbol struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement;index:idx_project_id,priority:2" json:"id"`
	ProjectID       uint64         `gorm:"not null;uniqueIndex:idx_project_uid,priority:1;index:idx_project_id,priority:1" json:"project_id"`
	UID             string         `gorm:"not null;size:255;uniqueIndex:idx_project_uid,priority:2" json:"uid"`
	Label           string         `gorm:"not null;size:255" json:"label"`
	ClassName       string         `gorm:"not null;size:255" json:"class_name"`
	ComponentTarget string         `gorm:"not null;size:255" json:"component_target"`
	Version         uint32         `gorm:"not null" json:"version"`
	SymbolData      *SymbolData    `gorm:"foreignKey:SymbolID;references:ID;constraint:OnDelete:CASCADE" json:"symbol_data,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Symbol) TableName() string {
	return "symbols"
}

type SymbolData struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SymbolID  uint64    `gorm:"not null;uniqueIndex" json:"symbol_id"`
	Data      *[]byte   `gorm:"not null;type:longblob" json:"data"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SymbolData) TableName() string {
	return "symbol_data"
}
