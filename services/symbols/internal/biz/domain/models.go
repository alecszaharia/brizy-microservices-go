// Package domain contains the business domain models and interfaces.
package domain

import "platform/pagination"

// SymbolData represents the data payload of a symbol.
type SymbolData struct {
	ID      uint64  `validate:"omitempty,gte=0"`
	Project uint64  `validate:"required,gt=0"`
	Data    *[]byte `validate:"omitempty"`
}

// ListSymbolsOptions contains parameters for listing symbols.
type ListSymbolsOptions struct {
	ProjectID  uint64                            `validate:"required,gt=0"`
	Pagination pagination.OffsetPaginationParams `validate:"required"`
}

// Symbol represents a symbol business entity.
type Symbol struct {
	ID              uint64      `validate:"omitempty,gte=0"`
	Project         uint64      `validate:"required,gt=0"`
	UID             string      `validate:"required,uuid4"`
	Label           string      `validate:"required,min=1,max=255"`
	ClassName       string      `validate:"required,min=1,max=255"`
	ComponentTarget string      `validate:"required,min=1,max=100"`
	Version         uint32      `validate:"required"`
	Data            *SymbolData `validate:"required"`
}
