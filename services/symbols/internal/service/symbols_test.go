package service

import (
	"context"
	v1 "contracts/gen/service/symbols/v1"
	"errors"
	"platform/pagination"
	"symbols/internal/biz/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSymbolUseCase struct {
	mock.Mock
}

func (uc *mockSymbolUseCase) GetSymbol(ctx context.Context, id uint64) (*domain.Symbol, error) {
	args := uc.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (uc *mockSymbolUseCase) CreateSymbol(ctx context.Context, g *domain.Symbol) (*domain.Symbol, error) {
	args := uc.Called(ctx, g)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (uc *mockSymbolUseCase) UpdateSymbol(ctx context.Context, g *domain.Symbol) (*domain.Symbol, error) {
	args := uc.Called(ctx, g)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (uc *mockSymbolUseCase) DeleteSymbol(ctx context.Context, id uint64) error {
	args := uc.Called(ctx, id)
	return args.Error(0)
}

func (uc *mockSymbolUseCase) ListSymbols(ctx context.Context, params *pagination.OffsetPaginationParams, filter map[string]interface{}) ([]*domain.Symbol, *pagination.Meta, error) {
	args := uc.Called(ctx, params, filter)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*domain.Symbol), nil, args.Error(2)
	}
	return args.Get(0).([]*domain.Symbol), args.Get(1).(*pagination.Meta), args.Error(2)
}

func TestCreateSymbol(t *testing.T) {
	tests := []struct {
		name        string
		request     *v1.CreateSymbolRequest
		mockSetup   func(*mockSymbolUseCase, context.Context, *v1.CreateSymbolRequest)
		wantErr     bool
		checkResult func(*testing.T, *v1.CreateSymbolResponse)
	}{
		{
			name: "success",
			request: &v1.CreateSymbolRequest{
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "component",
				Version:         1,
				Data:            []byte("test data"),
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.CreateSymbolRequest) {
				input := SymbolFromCreateRequest(req)
				output := &domain.Symbol{
					ID:              1,
					Project:         req.ProjectId,
					UID:             req.Uid,
					Label:           req.Label,
					ClassName:       req.ClassName,
					ComponentTarget: req.ComponentTarget,
					Version:         req.Version,
					Data: &domain.SymbolData{
						ID:      1,
						Project: req.ProjectId,
						Data:    &req.Data,
					},
				}
				uc.On("CreateSymbol", ctx, input).Return(output, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.CreateSymbolResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Symbol)
				assert.Equal(t, uint64(1), resp.Symbol.Id)
				assert.Equal(t, "Test Symbol", resp.Symbol.Label)
			},
		},
		{
			name: "use case error",
			request: &v1.CreateSymbolRequest{
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "component",
				Version:         1,
				Data:            []byte("test data"),
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.CreateSymbolRequest) {
				input := SymbolFromCreateRequest(req)
				uc.On("CreateSymbol", ctx, input).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success with empty data",
			request: &v1.CreateSymbolRequest{
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "component",
				Version:         1,
				Data:            []byte{},
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.CreateSymbolRequest) {
				input := SymbolFromCreateRequest(req)
				output := &domain.Symbol{
					ID:              1,
					Project:         req.ProjectId,
					UID:             req.Uid,
					Label:           req.Label,
					ClassName:       req.ClassName,
					ComponentTarget: req.ComponentTarget,
					Version:         req.Version,
					Data: &domain.SymbolData{
						ID:      1,
						Project: req.ProjectId,
						Data:    &req.Data,
					},
				}
				uc.On("CreateSymbol", ctx, input).Return(output, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.CreateSymbolResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Symbol)
				assert.Empty(t, resp.Symbol.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSymbolUseCase{}
			service := &SymbolService{uc: uc}
			ctx := context.Background()

			tt.mockSetup(uc, ctx, tt.request)

			result, err := service.CreateSymbol(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			uc.AssertExpectations(t)
		})
	}
}

func TestListSymbols(t *testing.T) {
	tests := []struct {
		name        string
		request     *v1.ListSymbolsRequest
		mockSetup   func(*mockSymbolUseCase, context.Context, *v1.ListSymbolsRequest)
		wantErr     bool
		checkResult func(*testing.T, *v1.ListSymbolsResponse)
	}{
		{
			name: "success with empty results",
			request: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Limit:     10,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.ListSymbolsRequest) {
				options, _ := NewListSymbolsOptions(req)
				meta := &pagination.Meta{
					TotalCount:      0,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				uc.On("ListSymbols", ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID}).Return([]*domain.Symbol{}, meta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.ListSymbolsResponse) {
				assert.NotNil(t, resp)
				assert.Empty(t, resp.Symbols)
				assert.NotNil(t, resp.Pagination)
				assert.Equal(t, uint64(0), resp.Pagination.TotalCount)
			},
		},
		{
			name: "success with multiple results",
			request: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Limit:     10,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.ListSymbolsRequest) {
				options, _ := NewListSymbolsOptions(req)
				data1 := []byte("data1")
				data2 := []byte("data2")
				symbols := []*domain.Symbol{
					{
						ID:              1,
						Project:         1,
						UID:             "uid1",
						Label:           "Symbol 1",
						ClassName:       "Class1",
						ComponentTarget: "component1",
						Version:         1,
						Data: &domain.SymbolData{
							ID:      1,
							Project: 1,
							Data:    &data1,
						},
					},
					{
						ID:              2,
						Project:         1,
						UID:             "uid2",
						Label:           "Symbol 2",
						ClassName:       "Class2",
						ComponentTarget: "component2",
						Version:         1,
						Data: &domain.SymbolData{
							ID:      2,
							Project: 1,
							Data:    &data2,
						},
					},
				}
				meta := &pagination.Meta{
					TotalCount:      2,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				uc.On("ListSymbols", ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID}).Return(symbols, meta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.ListSymbolsResponse) {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Symbols, 2)
				assert.Equal(t, "Symbol 1", resp.Symbols[0].Label)
				assert.Equal(t, "Symbol 2", resp.Symbols[1].Label)
				assert.NotNil(t, resp.Pagination)
				assert.Equal(t, uint64(2), resp.Pagination.TotalCount)
			},
		},
		{
			name: "success with pagination - second page",
			request: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    10,
				Limit:     10,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.ListSymbolsRequest) {
				options, _ := NewListSymbolsOptions(req)
				meta := &pagination.Meta{
					TotalCount:      25,
					Offset:          10,
					Limit:           10,
					HasNextPage:     true,
					HasPreviousPage: true,
				}
				uc.On("ListSymbols", ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID}).Return([]*domain.Symbol{}, meta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.ListSymbolsResponse) {
				assert.NotNil(t, resp)
				assert.Empty(t, resp.Symbols)
				assert.NotNil(t, resp.Pagination)
				assert.True(t, resp.Pagination.HasNextPage)
				assert.True(t, resp.Pagination.HasPreviousPage)
			},
		},
		{
			name: "use case error",
			request: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Limit:     10,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.ListSymbolsRequest) {
				options, _ := NewListSymbolsOptions(req)
				uc.On("ListSymbols", ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID}).Return(nil, nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success with nil symbol data",
			request: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Limit:     10,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.ListSymbolsRequest) {
				options, _ := NewListSymbolsOptions(req)
				symbols := []*domain.Symbol{
					{
						ID:              1,
						Project:         1,
						UID:             "uid1",
						Label:           "Symbol 1",
						ClassName:       "Class1",
						ComponentTarget: "component1",
						Version:         1,
						Data:            nil,
					},
				}
				meta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				uc.On("ListSymbols", ctx, &options.Pagination, map[string]interface{}{"project_id": options.ProjectID}).Return(symbols, meta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.ListSymbolsResponse) {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Symbols, 1)
				assert.IsType(t, &v1.SymbolItem{}, resp.Symbols[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSymbolUseCase{}
			service := &SymbolService{uc: uc}
			ctx := context.Background()

			tt.mockSetup(uc, ctx, tt.request)

			result, err := service.ListSymbols(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			uc.AssertExpectations(t)
		})
	}
}

func TestGetSymbol(t *testing.T) {
	tests := []struct {
		name        string
		request     *v1.GetSymbolRequest
		mockSetup   func(*mockSymbolUseCase, context.Context, *v1.GetSymbolRequest)
		wantErr     bool
		checkResult func(*testing.T, *v1.GetSymbolResponse)
	}{
		{
			name: "success",
			request: &v1.GetSymbolRequest{
				Id: 1,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.GetSymbolRequest) {
				data := []byte("test data")
				symbol := &domain.Symbol{
					ID:              1,
					Project:         1,
					UID:             "550e8400-e29b-41d4-a716-446655440000",
					Label:           "Test Symbol",
					ClassName:       "TestClass",
					ComponentTarget: "component",
					Version:         1,
					Data: &domain.SymbolData{
						ID:      1,
						Project: 1,
						Data:    &data,
					},
				}
				uc.On("GetSymbol", ctx, req.Id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.GetSymbolResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Symbol)
				assert.Equal(t, uint64(1), resp.Symbol.Id)
				assert.Equal(t, "Test Symbol", resp.Symbol.Label)
				assert.NotEmpty(t, resp.Symbol.Data)
			},
		},
		{
			name: "not found error",
			request: &v1.GetSymbolRequest{
				Id: 999,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.GetSymbolRequest) {
				uc.On("GetSymbol", ctx, req.Id).Return(nil, errors.New("symbol not found"))
			},
			wantErr: true,
		},
		{
			name: "success with nil data",
			request: &v1.GetSymbolRequest{
				Id: 1,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.GetSymbolRequest) {
				symbol := &domain.Symbol{
					ID:              1,
					Project:         1,
					UID:             "550e8400-e29b-41d4-a716-446655440000",
					Label:           "Test Symbol",
					ClassName:       "TestClass",
					ComponentTarget: "component",
					Version:         1,
					Data:            nil,
				}
				uc.On("GetSymbol", ctx, req.Id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.GetSymbolResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Symbol)
				assert.Nil(t, resp.Symbol.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSymbolUseCase{}
			service := &SymbolService{uc: uc}
			ctx := context.Background()

			tt.mockSetup(uc, ctx, tt.request)

			result, err := service.GetSymbol(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			uc.AssertExpectations(t)
		})
	}
}

func TestUpdateSymbol(t *testing.T) {
	tests := []struct {
		name        string
		request     *v1.UpdateSymbolRequest
		mockSetup   func(*mockSymbolUseCase, context.Context, *v1.UpdateSymbolRequest)
		wantErr     bool
		checkResult func(*testing.T, *v1.UpdateSymbolResponse)
	}{
		{
			name: "success",
			request: &v1.UpdateSymbolRequest{
				Id:              1,
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Updated Symbol",
				ClassName:       "UpdatedClass",
				ComponentTarget: "updated-component",
				Version:         1,
				Data:            []byte("updated data"),
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.UpdateSymbolRequest) {
				input := SymbolFromUpdateRequest(req)
				output := &domain.Symbol{
					ID:              req.Id,
					Project:         req.ProjectId,
					UID:             req.Uid,
					Label:           req.Label,
					ClassName:       req.ClassName,
					ComponentTarget: req.ComponentTarget,
					Version:         req.Version,
					Data: &domain.SymbolData{
						ID:      1,
						Project: req.ProjectId,
						Data:    &req.Data,
					},
				}
				uc.On("UpdateSymbol", ctx, input).Return(output, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *v1.UpdateSymbolResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Symbol)
				assert.Equal(t, uint64(1), resp.Symbol.Id)
				assert.Equal(t, "Updated Symbol", resp.Symbol.Label)
				assert.Equal(t, uint32(1), resp.Symbol.Version)
			},
		},
		{
			name: "not found error",
			request: &v1.UpdateSymbolRequest{
				Id:              999,
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Updated Symbol",
				ClassName:       "UpdatedClass",
				ComponentTarget: "updated-component",
				Version:         1,
				Data:            []byte("updated data"),
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.UpdateSymbolRequest) {
				input := SymbolFromUpdateRequest(req)
				uc.On("UpdateSymbol", ctx, input).Return(nil, errors.New("symbol not found"))
			},
			wantErr: true,
		},
		{
			name: "validation error",
			request: &v1.UpdateSymbolRequest{
				Id:              1,
				ProjectId:       1,
				Uid:             "invalid-uid",
				Label:           "Updated Symbol",
				ClassName:       "UpdatedClass",
				ComponentTarget: "updated-component",
				Version:         1,
				Data:            []byte("updated data"),
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.UpdateSymbolRequest) {
				input := SymbolFromUpdateRequest(req)
				uc.On("UpdateSymbol", ctx, input).Return(nil, errors.New("validation error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSymbolUseCase{}
			service := &SymbolService{uc: uc}
			ctx := context.Background()

			tt.mockSetup(uc, ctx, tt.request)

			result, err := service.UpdateSymbol(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			uc.AssertExpectations(t)
		})
	}
}

func TestDeleteSymbol(t *testing.T) {
	tests := []struct {
		name      string
		request   *v1.DeleteSymbolRequest
		mockSetup func(*mockSymbolUseCase, context.Context, *v1.DeleteSymbolRequest)
		wantErr   bool
	}{
		{
			name: "success",
			request: &v1.DeleteSymbolRequest{
				Id: 1,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.DeleteSymbolRequest) {
				uc.On("DeleteSymbol", ctx, req.Id).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			request: &v1.DeleteSymbolRequest{
				Id: 999,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.DeleteSymbolRequest) {
				uc.On("DeleteSymbol", ctx, req.Id).Return(errors.New("symbol not found"))
			},
			wantErr: true,
		},
		{
			name: "database error",
			request: &v1.DeleteSymbolRequest{
				Id: 1,
			},
			mockSetup: func(uc *mockSymbolUseCase, ctx context.Context, req *v1.DeleteSymbolRequest) {
				uc.On("DeleteSymbol", ctx, req.Id).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSymbolUseCase{}
			service := &SymbolService{uc: uc}
			ctx := context.Background()

			tt.mockSetup(uc, ctx, tt.request)

			result, err := service.DeleteSymbol(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			uc.AssertExpectations(t)
		})
	}
}
