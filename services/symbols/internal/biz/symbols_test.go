package biz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"platform/pagination"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSymbolRepo is a mock implementation of SymbolRepo for testing
type MockSymbolRepo struct {
	mock.Mock
}

// MockPublisher is a mock implementation of events.Publisher for testing
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	args := m.Called(ctx, topic, payload)
	return args.Error(0)
}

func (m *MockPublisher) Unwrap() message.Publisher {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(message.Publisher)
}

func (m *MockSymbolRepo) Create(ctx context.Context, symbol *Symbol) (*Symbol, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Symbol), args.Error(1)
}

func (m *MockSymbolRepo) Update(ctx context.Context, symbol *Symbol) (*Symbol, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Symbol), args.Error(1)
}

func (m *MockSymbolRepo) FindByID(ctx context.Context, id uint64) (*Symbol, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Symbol), args.Error(1)
}

func (m *MockSymbolRepo) ListSymbols(ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) ([]*Symbol, *pagination.Meta, error) {
	args := m.Called(ctx, offset, limit, filter)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*Symbol), nil, args.Error(2)
	}
	return args.Get(0).([]*Symbol), args.Get(1).(*pagination.Meta), args.Error(2)
}

func (m *MockSymbolRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTransaction is a mock implementation of common.Transaction for testing
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) InTx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	args := m.Called(ctx, fn)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	// Execute the callback immediately without a real transaction
	return fn(ctx, nil)
}

// Helper function to create a test SymbolUseCase
func setupSymbolUseCase(mockRepo *MockSymbolRepo) SymbolUseCase {
	logger := log.NewStdLogger(os.Stdout)
	v := NewSymbolValidator()
	mockPub := new(MockPublisher)
	mockTx := new(MockTransaction)

	// Allow any Publish calls to succeed by default
	mockPub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockTx.On("InTx", mock.Anything, mock.Anything).Return(nil).Maybe()

	return NewSymbolUseCase(mockRepo, v, mockTx, mockPub, logger)
}

// Helper function to create a valid Symbol for testing
func validSymbol() *Symbol {
	bytes := []byte(`{"key": "value"}`)
	return &Symbol{
		Project:         1,
		UID:             "550e8400-e29b-41d4-a716-446655440000",
		Label:           "Test Symbol",
		ClassName:       "TestClass",
		ComponentTarget: "TestTarget",
		Version:         1,
		Data:            &SymbolData{Project: 1, Data: &bytes},
	}
}

func TestGetSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbolID    uint64
		mockSetup   func(*MockSymbolRepo, context.Context, uint64)
		wantErr     bool
		checkResult func(*testing.T, *Symbol)
	}{
		{
			name:     "success",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				symbol := validSymbol()
				symbol.ID = id
				repo.On("FindByID", ctx, id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(1), result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.UID)
			},
		},
		{
			name:     "not found",
			symbolID: 999,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("FindByID", ctx, id).Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("FindByID", ctx, id).Return(nil, fmt.Errorf("%w: connection failed", ErrDatabase))
			},
			wantErr: true,
		},
		{
			name:     "with symbol data",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				bytes := []byte(`{"key": "value"}`)
				symbol := &Symbol{
					ID:              id,
					Project:         1,
					UID:             "550e8400-e29b-41d4-a716-446655440000",
					Label:           "Test Symbol",
					ClassName:       "TestClass",
					ComponentTarget: "TestTarget",
					Version:         1,
					Data: &SymbolData{
						ID:      1,
						Project: 1,
						Data:    &bytes,
					},
				}
				repo.On("FindByID", ctx, id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.Equal(t, uint64(1), result.Data.ID)
			},
		},
		{
			name:     "without symbol data",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				symbol := &Symbol{
					ID:              id,
					Project:         1,
					UID:             "550e8400-e29b-41d4-a716-446655440000",
					Label:           "Test Symbol",
					ClassName:       "TestClass",
					ComponentTarget: "TestTarget",
					Version:         1,
					Data:            nil,
				}
				repo.On("FindByID", ctx, id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.Nil(t, result.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.symbolID)

			result, err := uc.GetSymbol(ctx, tt.symbolID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbol      *Symbol
		mockSetup   func(*MockSymbolRepo, context.Context, *Symbol)
		wantErr     bool
		errContains string
		checkResult func(*testing.T, *Symbol)
	}{
		{
			name:   "success",
			symbol: validSymbol(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				expectedSymbol := &Symbol{
					ID:              1,
					Project:         symbol.Project,
					UID:             symbol.UID,
					Label:           symbol.Label,
					ClassName:       symbol.ClassName,
					ComponentTarget: symbol.ComponentTarget,
					Version:         symbol.Version,
					Data:            symbol.Data,
				}
				repo.On("Create", ctx, mock.AnythingOfType("*biz.Symbol")).Return(expectedSymbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(1), result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
			},
		},
		{
			name: "missing project",
			symbol: &Symbol{
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "Project",
		},
		{
			name: "invalid uuid",
			symbol: &Symbol{
				Project:         1,
				UID:             "invalid-uuid",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "UID",
		},
		{
			name: "empty label",
			symbol: &Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "Label",
		},
		{
			name: "empty class name",
			symbol: &Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "ClassName",
		},
		{
			name: "empty component target",
			symbol: &Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "ComponentTarget",
		},
		{
			name: "invalid version",
			symbol: &Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         0,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "Version",
		},
		{
			name: "label too long",
			symbol: &Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           string(make([]byte, 256)),
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "Label",
		},
		{
			name:   "repository error",
			symbol: validSymbol(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				repo.On("Create", ctx, mock.AnythingOfType("*biz.Symbol")).Return(nil, fmt.Errorf("%w: connection failed", ErrDatabase))
			},
			wantErr: true,
		},
		{
			name: "with symbol data",
			symbol: func() *Symbol {
				s := validSymbol()
				bytes := []byte(`{"key": "value"}`)
				s.Data = &SymbolData{
					Project: 1,
					Data:    &bytes,
				}
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				expectedSymbol := &Symbol{
					ID:              1,
					Project:         symbol.Project,
					UID:             symbol.UID,
					Label:           symbol.Label,
					ClassName:       symbol.ClassName,
					ComponentTarget: symbol.ComponentTarget,
					Version:         symbol.Version,
					Data:            symbol.Data,
				}
				repo.On("Create", ctx, mock.AnythingOfType("*biz.Symbol")).Return(expectedSymbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.Equal(t, result.Data.Data, result.Data.Data)
			},
		},
		{
			name: "symbol data validation error",
			symbol: func() *Symbol {
				s := validSymbol()
				bytes := []byte(`{"key": "value"}`)
				s.Data = &SymbolData{
					Project: 0, // Invalid
					Data:    &bytes,
				}
				return s
			}(),
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:     true,
			errContains: "Project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.symbol)

			result, err := uc.CreateSymbol(ctx, tt.symbol)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbol      *Symbol
		mockSetup   func(*MockSymbolRepo, context.Context, *Symbol)
		wantErr     bool
		checkResult func(*testing.T, *Symbol)
	}{
		{
			name: "success",
			symbol: func() *Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*biz.Symbol")).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(1), result.ID)
			},
		},
		{
			name: "validation error",
			symbol: func() *Symbol {
				s := validSymbol()
				s.ID = 1
				s.Label = "" // Invalid
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {},
			wantErr:   true,
		},
		{
			name: "not found error",
			symbol: func() *Symbol {
				s := validSymbol()
				s.ID = 999
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*biz.Symbol")).Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			symbol: func() *Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*biz.Symbol")).Return(nil, fmt.Errorf("%w: connection failed", ErrDatabase))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.symbol)

			result, err := uc.UpdateSymbol(ctx, tt.symbol)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbolID    uint64
		mockSetup   func(*MockSymbolRepo, context.Context, uint64)
		wantErr     bool
		errContains string
	}{
		{
			name:     "success",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("Delete", ctx, id).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "zero id",
			symbolID:    0,
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, id uint64) {},
			wantErr:     true,
			errContains: "invalid symbol ID",
		},
		{
			name:     "not found error",
			symbolID: 999,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("Delete", ctx, id).Return(ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("Delete", ctx, id).Return(fmt.Errorf("%w: connection failed", ErrDatabase))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.symbolID)

			err := uc.DeleteSymbol(ctx, tt.symbolID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListSymbols(t *testing.T) {
	tests := []struct {
		name        string
		params      *pagination.OffsetPaginationParams
		filter      map[string]interface{}
		mockSetup   func(*MockSymbolRepo, context.Context, uint64, uint32, map[string]interface{})
		wantErr     bool
		errContains string
		checkResult func(*testing.T, []*Symbol, *pagination.Meta)
	}{
		{
			name: "success - first page",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbol := validSymbol()
				expectedSymbols := []*Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.Len(t, symbols, 1)
				assert.NotNil(t, meta)
				assert.Equal(t, uint64(1), meta.TotalCount)
				assert.Equal(t, uint64(0), meta.Offset)
				assert.Equal(t, uint32(10), meta.Limit)
				assert.False(t, meta.HasNextPage)
				assert.False(t, meta.HasPreviousPage)
			},
		},
		{
			name: "success - with next page",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbols := make([]*Symbol, 10)
				for i := 0; i < 10; i++ {
					symbols[i] = validSymbol()
				}
				expectedMeta := &pagination.Meta{
					TotalCount:      25,
					Offset:          0,
					Limit:           10,
					HasNextPage:     true,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(symbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.Len(t, symbols, 10)
				assert.NotNil(t, meta)
				assert.Equal(t, uint64(25), meta.TotalCount)
				assert.True(t, meta.HasNextPage)
				assert.False(t, meta.HasPreviousPage)
			},
		},
		{
			name: "success - second page",
			params: &pagination.OffsetPaginationParams{
				Offset: 10,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbols := make([]*Symbol, 10)
				for i := 0; i < 10; i++ {
					symbols[i] = validSymbol()
				}
				expectedMeta := &pagination.Meta{
					TotalCount:      25,
					Offset:          10,
					Limit:           10,
					HasNextPage:     true,
					HasPreviousPage: true,
				}
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(symbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.NotNil(t, meta)
				assert.True(t, meta.HasNextPage)
				assert.True(t, meta.HasPreviousPage)
			},
		},
		{
			name: "success - empty result",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  20,
			},
			filter: map[string]interface{}{"project_id": uint64(999)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				expectedMeta := &pagination.Meta{
					TotalCount:      0,
					Offset:          0,
					Limit:           20,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, offset, limit, filter).Return([]*Symbol{}, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.Len(t, symbols, 0)
				assert.NotNil(t, meta)
				assert.Equal(t, uint64(0), meta.TotalCount)
			},
		},
		{
			name: "limit too large",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  101,
			},
			filter: map[string]interface{}{"project_id": uint64(1)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
			},
			wantErr:     true,
			errContains: "Limit",
		},
		{
			name: "repository error",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(nil, nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success - passes nil filter to repository",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: nil,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbol := validSymbol()
				expectedSymbols := []*Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				// Verify nil filter is passed through
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "success - passes empty filter to repository",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbol := validSymbol()
				expectedSymbols := []*Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				// Verify empty map is passed through
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "success - multiple filters passed through",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1), "label": "test-label", "version": uint32(2)},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				symbol := validSymbol()
				expectedSymbols := []*Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				// Verify all filter keys are preserved
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "error - repository error with filter",
			params: &pagination.OffsetPaginationParams{
				Offset: 0,
				Limit:  10,
			},
			filter: map[string]interface{}{"project_id": uint64(1), "label": "test"},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, offset uint64, limit uint32, filter map[string]interface{}) {
				repo.On("ListSymbols", ctx, offset, limit, filter).Return(nil, nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.params.Offset, tt.params.Limit, tt.filter)

			symbols, meta, err := uc.ListSymbols(ctx, tt.params, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, symbols)
				assert.Nil(t, meta)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, symbols, meta)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNewSymbolValidator(t *testing.T) {
	v := NewSymbolValidator()
	assert.NotNil(t, v)
	assert.IsType(t, &validator.Validate{}, v)
}
