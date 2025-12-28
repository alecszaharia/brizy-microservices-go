package pagination

// OffsetPaginationParams contains offset/limit parameters for paginated queries
type OffsetPaginationParams struct {
	Offset uint64 `validate:"gte=0"`
	Limit  uint32 `validate:"gte=1,lte=100"`
}

// PaginationMeta contains metadata about paginated results
type Meta struct {
	TotalCount      uint64
	Offset          uint64
	Limit           uint32
	HasNextPage     bool
	HasPreviousPage bool
}
