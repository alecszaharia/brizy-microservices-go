package biz

import "platform/pagination"

type SymbolData struct {
	Id      uint64  `validate:"omitempty,gte=0"`
	Project uint64  `validate:"required,gt=0"`
	Data    *[]byte `validate:"omitempty"`
}

// ListSymbolsOptions contains parameters for listing symbols
type ListSymbolsOptions struct {
	ProjectID  uint64                      `validate:"required,gt=0"`
	Pagination pagination.PaginationParams `validate:"required"`
}

type Symbol struct {
	Id              uint64      `validate:"omitempty,gte=0"`
	Project         uint64      `validate:"required,gt=0"`
	Uid             string      `validate:"required,uuid4"`
	Label           string      `validate:"required,min=1,max=255"`
	ClassName       string      `validate:"required,min=1,max=255"`
	ComponentTarget string      `validate:"required,min=1,max=100"`
	Version         uint32      `validate:"required"`
	Data            *SymbolData `validate:"required"`
}
