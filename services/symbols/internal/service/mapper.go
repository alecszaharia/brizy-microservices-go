// Package service implements the service layer with gRPC/HTTP handlers.
package service

import (
	v1 "contracts/gen/service/symbols/v1"
	"fmt"
	"platform/pagination"

	"symbols/internal/biz/domain"

	"github.com/go-kratos/kratos/v2/errors"
)

func toBizSymbol(s *v1.Symbol) *domain.Symbol {
	return &domain.Symbol{
		ID:              s.Id,
		Project:         s.ProjectId,
		UID:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data: &domain.SymbolData{
			Project: s.ProjectId,
			Data:    &s.Data,
		},
	}
}

func SymbolFromCreateRequest(s *v1.CreateSymbolRequest) *domain.Symbol {
	return &domain.Symbol{
		Project:         s.ProjectId,
		UID:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data: &domain.SymbolData{
			Project: s.ProjectId,
			Data:    &s.Data,
		},
	}
}
func SymbolFromUpdateRequest(s *v1.UpdateSymbolRequest) *domain.Symbol {
	return &domain.Symbol{
		ID:              s.Id,
		Project:         s.ProjectId,
		UID:             s.Uid,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data: &domain.SymbolData{
			Project: s.ProjectId,
			Data:    &s.Data,
		},
	}
}

// NewListSymbolsOptions transforms proto request to domain options with defaults
func NewListSymbolsOptions(in *v1.ListSymbolsRequest) (*domain.ListSymbolsOptions, error) {
	// Apply default values if not provided
	offset := in.Offset
	limit := in.Limit

	// Default limit to 20 if not provided or zero
	if limit == 0 {
		limit = 20
	}

	options := &domain.ListSymbolsOptions{
		ProjectID: in.ProjectId,
		Pagination: pagination.OffsetPaginationParams{
			Offset: offset,
			Limit:  limit,
		},
	}

	return options, nil
}

func toV1Symbol(s *domain.Symbol) *v1.Symbol {
	var data []byte
	if s.Data != nil && s.Data.Data != nil {
		data = *s.Data.Data
	}
	return &v1.Symbol{
		Id:              s.ID,
		ProjectId:       s.Project,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		Data:            data,
	}
}
func toV1SymbolItem(s *domain.Symbol) *v1.SymbolItem {
	return &v1.SymbolItem{
		Id:              s.ID,
		ProjectId:       s.Project,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
	}
}

// toV1PaginationMeta transforms domain pagination metadata to proto metadata
func toV1PaginationMeta(meta *pagination.Meta) *v1.PaginationMeta {
	if meta == nil {
		return nil
	}

	return &v1.PaginationMeta{
		TotalCount:      meta.TotalCount,
		Offset:          meta.Offset,
		Limit:           meta.Limit,
		HasNextPage:     meta.HasNextPage,
		HasPreviousPage: meta.HasPreviousPage,
	}
}

func toServiceError(err error) *errors.Error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrSymbolNotFound):
		return errors.NotFound(
			v1.ErrorReason_SYMBOL_NOT_FOUND.String(),
			"symbol not found",
		)

	case errors.Is(err, domain.ErrInvalidID):
		return errors.BadRequest(
			v1.ErrorReason_INVALID_ID.String(),
			"invalid symbol ID: must be greater than zero",
		)

	case errors.Is(err, domain.ErrDuplicateSymbol):
		return errors.BadRequest(
			v1.ErrorReason_DUPLICATE_SYMBOL.String(),
			"symbol with this UID already exists in the project",
		)

	case errors.Is(err, domain.ErrValidationFailed):
		return errors.BadRequest(
			v1.ErrorReason_VALIDATION_ERROR.String(),
			fmt.Sprintf("validation failed: %v", err),
		)

	case errors.Is(err, domain.ErrDatabaseOperation):
		return errors.InternalServer(
			v1.ErrorReason_DATABASE_ERROR.String(),
			"database operation failed",
		)

	default:
		// Unknown error - return as internal server error
		return errors.InternalServer(
			v1.ErrorReason_SYMBOL_UNSPECIFIED.String(),
			"internal server error",
		).WithCause(err)
	}

}
