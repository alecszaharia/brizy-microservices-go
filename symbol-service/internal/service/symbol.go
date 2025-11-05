package service

import (
	"context"

	v1 "symbol-service/api/symbol/v1"
	"symbol-service/internal/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SymbolService struct {
	v1.UnimplementedSymbolsServer

	uc *biz.SymbolUseCase
}

func NewSymbolService(uc *biz.SymbolUseCase) *SymbolService {
	return &SymbolService{uc: uc}
}

func (s *SymbolService) CreateSymbol(ctx context.Context, in *v1.CreateSymbolRequest) (*v1.CreateSymbolResponse, error) {
	g, err := s.uc.CreateSymbol(ctx, toBizSymbolFromRequest(in))
	if err != nil {
		return nil, err
	}
	return &v1.CreateSymbolResponse{Symbol: toV1Symbol(g)}, nil
}
func (s *SymbolService) ListSymbols(ctx context.Context, in *v1.ListSymbolsRequest) (*v1.ListSymbolsResponse, error) {
	symbols, cursor, err := s.uc.ListSymbols(ctx, toBizListSymbolsOptions(in))
	if err != nil {
		return nil, err
	}

	var result []*v1.Symbol

	for _, symbol := range symbols {
		result = append(result, toV1Symbol(symbol))
	}

	return &v1.ListSymbolsResponse{Symbols: result, PageToken: toV1PageToken(cursor)}, nil
}
func (s *SymbolService) GetSymbol(context.Context, *v1.GetSymbolRequest) (*v1.GetSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSymbol not implemented")
}
func (s *SymbolService) UpdateSymbol(context.Context, *v1.UpdateSymbolRequest) (*v1.UpdateSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSymbol not implemented")
}
func (s *SymbolService) DeleteSymbol(context.Context, *v1.DeleteSymbolRequest) (*v1.DeleteSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSymbol not implemented")
}
