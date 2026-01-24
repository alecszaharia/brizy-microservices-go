// Package domain contains the business domain models and interfaces.
package domain

import "platform/pagination"

// SortDirection represents the direction of sorting.
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// SortField represents valid fields for sorting symbols.
type SortField string

const (
	SortByID        SortField = "id"
	SortByCreatedAt SortField = "created_at"
	SortByUpdatedAt SortField = "updated_at"
	SortByLabel     SortField = "label"
)

// IsValid checks if the sort field is valid.
func (sf SortField) IsValid() bool {
	switch sf {
	case SortByID, SortByCreatedAt, SortByUpdatedAt, SortByLabel:
		return true
	}
	return false
}

// SortOption represents a sorting configuration.
type SortOption struct {
	Field     SortField
	Direction SortDirection
}

// DefaultSortOption returns deterministic default (ID ASC).
func DefaultSortOption() SortOption {
	return SortOption{Field: SortByID, Direction: SortAsc}
}

// SymbolData represents the data payload of a symbol.
type SymbolData struct {
	ID      uint64  `validate:"omitempty,gte=0"`
	Project uint64  `validate:"required,gt=0"`
	Data    *[]byte `validate:"omitempty"`
}

// SymbolFilter contains optional filter criteria for listing symbols.
// All optional fields are pointers to distinguish between "not set" and "empty value".
type SymbolFilter struct {
	ProjectID       uint64  // Required, always set
	Label           *string // Optional: exact match on label
	ComponentTarget *string // Optional: exact match on component_target
}

// ListSymbolsOptions contains parameters for listing symbols.
type ListSymbolsOptions struct {
	Filter     SymbolFilter                      `validate:"required"`
	Pagination pagination.OffsetPaginationParams `validate:"required"`
	Sort       SortOption                        // Sorting options (defaults to ID ASC if empty)
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
