package service

import (
	"context"
	v1 "contracts/gen/service/symbols/v1"

	"symbols/internal/biz/domain"
)

type SymbolService struct {
	v1.UnimplementedSymbolsServiceServer

	uc domain.SymbolUseCase
}

func NewSymbolService(uc domain.SymbolUseCase) *SymbolService {
	return &SymbolService{uc: uc}
}

func (s *SymbolService) CreateSymbol(ctx context.Context, in *v1.CreateSymbolRequest) (*v1.CreateSymbolResponse, error) {
	symbol := SymbolFromCreateRequest(in)
	g, err := s.uc.CreateSymbol(ctx, symbol)
	if err != nil {
		return nil, toServiceError(err)
	}
	return &v1.CreateSymbolResponse{Symbol: toV1Symbol(g)}, nil
}
func (s *SymbolService) ListSymbols(ctx context.Context, in *v1.ListSymbolsRequest) (*v1.ListSymbolsResponse, error) {
	// Transform request to domain options (applies defaults)
	options, err := NewListSymbolsOptions(in)
	if err != nil {
		return nil, toServiceError(err)
	}

	// Call business layer to get symbols and pagination metadata
	symbols, meta, err := s.uc.ListSymbols(ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID})
	if err != nil {
		return nil, toServiceError(err)
	}

	// Transform symbols to proto format
	result := make([]*v1.SymbolItem, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, toV1SymbolItem(symbol))
	}

	// Transform pagination metadata to proto format
	paginationMeta := toV1PaginationMeta(meta)

	return &v1.ListSymbolsResponse{
		Symbols:    result,
		Pagination: paginationMeta,
	}, nil
}
func (s *SymbolService) GetSymbol(ctx context.Context, in *v1.GetSymbolRequest) (*v1.GetSymbolResponse, error) {
	g, err := s.uc.GetSymbol(ctx, in.Id)
	if err != nil {
		return nil, toServiceError(err)
	}
	return &v1.GetSymbolResponse{Symbol: toV1Symbol(g)}, nil
}
func (s *SymbolService) UpdateSymbol(ctx context.Context, in *v1.UpdateSymbolRequest) (*v1.UpdateSymbolResponse, error) {
	symbol := SymbolFromUpdateRequest(in)
	g, err := s.uc.UpdateSymbol(ctx, symbol)
	if err != nil {
		return nil, toServiceError(err)
	}
	return &v1.UpdateSymbolResponse{Symbol: toV1Symbol(g)}, nil
}
func (s *SymbolService) DeleteSymbol(ctx context.Context, in *v1.DeleteSymbolRequest) (*v1.DeleteSymbolResponse, error) {
	if err := s.uc.DeleteSymbol(ctx, in.Id); err != nil {
		return nil, toServiceError(err)
	}
	return &v1.DeleteSymbolResponse{Success: true}, nil
}
