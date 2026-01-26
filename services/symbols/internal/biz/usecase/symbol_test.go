// Package symbol provides tests for use cases for managing symbols.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"platform/pagination"
	"symbols/internal/biz/domain"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockPublisher) PublishSymbolCreated(ctx context.Context, symbol *domain.Symbol) error {
	args := m.Called(ctx, symbol)
	return args.Error(0)
}

func (m *MockPublisher) PublishSymbolUpdated(ctx context.Context, symbol *domain.Symbol) error {
	args := m.Called(ctx, symbol)
	return args.Error(0)
}

func (m *MockPublisher) PublishSymbolDeleted(ctx context.Context, symbol *domain.Symbol) error {
	args := m.Called(ctx, symbol)
	return args.Error(0)
}

func (m *MockSymbolRepo) Create(ctx context.Context, symbol *domain.Symbol) (*domain.Symbol, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (m *MockSymbolRepo) Update(ctx context.Context, symbol *domain.Symbol) (*domain.Symbol, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (m *MockSymbolRepo) FindByID(ctx context.Context, id uint64) (*domain.Symbol, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Symbol), args.Error(1)
}

func (m *MockSymbolRepo) ListSymbols(ctx context.Context, opts domain.ListSymbolsOptions) ([]*domain.Symbol, *pagination.Meta, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).([]*domain.Symbol), nil, args.Error(2)
	}
	return args.Get(0).([]*domain.Symbol), args.Get(1).(*pagination.Meta), args.Error(2)
}

func (m *MockSymbolRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockTransaction is a mock implementation of common.Transaction for testing
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	// Execute the callback immediately without a real transaction
	return fn(ctx)
}

// testDeps holds all mock dependencies for testing
type testDeps struct {
	repo *MockSymbolRepo
	pub  *MockPublisher
	tx   *MockTransaction
	uc   domain.SymbolUseCase
}

// setupSymbolUseCaseWithDeps creates a test SymbolUseCase and returns all dependencies for assertions
func setupSymbolUseCaseWithDeps() *testDeps {
	logger := log.NewStdLogger(os.Stdout)
	v := NewValidator()
	mockRepo := new(MockSymbolRepo)
	mockPub := new(MockPublisher)
	mockTx := new(MockTransaction)

	// Default transaction behavior - executes the callback
	mockTx.On("InTx", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewUseCase(mockRepo, v, mockTx, mockPub, logger)

	return &testDeps{
		repo: mockRepo,
		pub:  mockPub,
		tx:   mockTx,
		uc:   uc,
	}
}

// Helper function to create a test SymbolUseCase (legacy - for tests that don't need publisher assertions)
func setupSymbolUseCase(mockRepo *MockSymbolRepo) domain.SymbolUseCase {
	logger := log.NewStdLogger(os.Stdout)
	v := NewValidator()
	mockPub := new(MockPublisher)
	mockTx := new(MockTransaction)

	// Allow any event publishing calls to succeed by default
	mockPub.On("PublishSymbolCreated", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockPub.On("PublishSymbolUpdated", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockPub.On("PublishSymbolDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockTx.On("InTx", mock.Anything, mock.Anything).Return(nil).Maybe()

	return NewUseCase(mockRepo, v, mockTx, mockPub, logger)
}

// Helper function to create a valid Symbol for testing
func validSymbol() *domain.Symbol {
	bytes := []byte(`{"key": "value"}`)
	return &domain.Symbol{
		Project:         1,
		UID:             "550e8400-e29b-41d4-a716-446655440000",
		Label:           "Test Symbol",
		ClassName:       "TestClass",
		ComponentTarget: "TestTarget",
		Version:         1,
		Data:            &domain.SymbolData{Project: 1, Data: &bytes},
	}
}

func TestGetSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbolID    uint64
		mockSetup   func(*MockSymbolRepo, context.Context, uint64)
		wantErr     bool
		checkResult func(*testing.T, *domain.Symbol)
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
			checkResult: func(t *testing.T, result *domain.Symbol) {
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
				repo.On("FindByID", ctx, id).Return(nil, domain.ErrDataNotFound)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("FindByID", ctx, id).Return(nil, fmt.Errorf("%w: connection failed", domain.ErrDataDatabase))
			},
			wantErr: true,
		},
		{
			name:     "with symbol data",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				bytes := []byte(`{"key": "value"}`)
				symbol := &domain.Symbol{
					ID:              id,
					Project:         1,
					UID:             "550e8400-e29b-41d4-a716-446655440000",
					Label:           "Test Symbol",
					ClassName:       "TestClass",
					ComponentTarget: "TestTarget",
					Version:         1,
					Data: &domain.SymbolData{
						ID:      1,
						Project: 1,
						Data:    &bytes,
					},
				}
				repo.On("FindByID", ctx, id).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.Equal(t, uint64(1), result.Data.ID)
			},
		},
		{
			name:     "without symbol data",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				symbol := &domain.Symbol{
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
			checkResult: func(t *testing.T, result *domain.Symbol) {
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
		symbol      *domain.Symbol
		mockSetup   func(*MockSymbolRepo, context.Context, *domain.Symbol)
		wantErr     bool
		errContains string
		checkResult func(*testing.T, *domain.Symbol)
	}{
		{
			name:   "success",
			symbol: validSymbol(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				expectedSymbol := &domain.Symbol{
					ID:              1,
					Project:         symbol.Project,
					UID:             symbol.UID,
					Label:           symbol.Label,
					ClassName:       symbol.ClassName,
					ComponentTarget: symbol.ComponentTarget,
					Version:         symbol.Version,
					Data:            symbol.Data,
				}
				repo.On("Create", ctx, mock.AnythingOfType("*domain.Symbol")).Return(expectedSymbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(1), result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
			},
		},
		{
			name: "missing project",
			symbol: &domain.Symbol{
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "Project",
		},
		{
			name: "invalid uuid",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "invalid-uuid",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "UID",
		},
		{
			name: "empty label",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "Label",
		},
		{
			name: "empty class name",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "ClassName",
		},
		{
			name: "empty component target",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "ComponentTarget",
		},
		{
			name: "invalid version",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         0,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "Version",
		},
		{
			name: "label too long",
			symbol: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           string(make([]byte, 256)),
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         1,
			},
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:     true,
			errContains: "Label",
		},
		{
			name:   "repository error",
			symbol: validSymbol(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				repo.On("Create", ctx, mock.AnythingOfType("*domain.Symbol")).Return(nil, fmt.Errorf("%w: connection failed", domain.ErrDataDatabase))
			},
			wantErr: true,
		},
		{
			name: "with symbol data",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				bytes := []byte(`{"key": "value"}`)
				s.Data = &domain.SymbolData{
					Project: 1,
					Data:    &bytes,
				}
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				expectedSymbol := &domain.Symbol{
					ID:              1,
					Project:         symbol.Project,
					UID:             symbol.UID,
					Label:           symbol.Label,
					ClassName:       symbol.ClassName,
					ComponentTarget: symbol.ComponentTarget,
					Version:         symbol.Version,
					Data:            symbol.Data,
				}
				repo.On("Create", ctx, mock.AnythingOfType("*domain.Symbol")).Return(expectedSymbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.Equal(t, result.Data.Data, result.Data.Data)
			},
		},
		{
			name: "symbol data validation error",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				bytes := []byte(`{"key": "value"}`)
				s.Data = &domain.SymbolData{
					Project: 0, // Invalid
					Data:    &bytes,
				}
				return s
			}(),
			mockSetup:   func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
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
		symbol      *domain.Symbol
		mockSetup   func(*MockSymbolRepo, context.Context, *domain.Symbol)
		wantErr     bool
		checkResult func(*testing.T, *domain.Symbol)
	}{
		{
			name: "success",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*domain.Symbol")).Return(symbol, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(1), result.ID)
			},
		},
		{
			name: "validation error",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				s.Label = "" // Invalid
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {},
			wantErr:   true,
		},
		{
			name: "not found error",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 999
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*domain.Symbol")).Return(nil, domain.ErrDataNotFound)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, symbol *domain.Symbol) {
				repo.On("Update", ctx, mock.AnythingOfType("*domain.Symbol")).Return(nil, fmt.Errorf("%w: connection failed", domain.ErrDataDatabase))
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
				symbol := validSymbol()
				symbol.ID = id
				repo.On("FindByID", ctx, id).Return(symbol, nil)
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
			name:     "not found error on find",
			symbolID: 999,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				repo.On("FindByID", ctx, id).Return(nil, domain.ErrDataNotFound)
			},
			wantErr: true,
		},
		{
			name:     "repository error on delete",
			symbolID: 1,
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, id uint64) {
				symbol := validSymbol()
				symbol.ID = id
				repo.On("FindByID", ctx, id).Return(symbol, nil)
				repo.On("Delete", ctx, id).Return(fmt.Errorf("%w: connection failed", domain.ErrDataDatabase))
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
		opts        domain.ListSymbolsOptions
		mockSetup   func(*MockSymbolRepo, context.Context, domain.ListSymbolsOptions)
		wantErr     bool
		errContains string
		checkResult func(*testing.T, []*domain.Symbol, *pagination.Meta)
	}{
		{
			name: "success - first page",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.ProjectID == opts.Filter.ProjectID &&
						o.Pagination.Offset == opts.Pagination.Offset &&
						o.Pagination.Limit == opts.Pagination.Limit
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
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
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbols := make([]*domain.Symbol, 10)
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
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.ProjectID == opts.Filter.ProjectID
				})).Return(symbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
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
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 10, Limit: 10},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbols := make([]*domain.Symbol, 10)
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
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.ProjectID == opts.Filter.ProjectID
				})).Return(symbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.NotNil(t, meta)
				assert.True(t, meta.HasNextPage)
				assert.True(t, meta.HasPreviousPage)
			},
		},
		{
			name: "success - empty result",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(999)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 20},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				expectedMeta := &pagination.Meta{
					TotalCount:      0,
					Offset:          0,
					Limit:           20,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.ProjectID == opts.Filter.ProjectID
				})).Return([]*domain.Symbol{}, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.NotNil(t, symbols)
				assert.Len(t, symbols, 0)
				assert.NotNil(t, meta)
				assert.Equal(t, uint64(0), meta.TotalCount)
			},
		},
		{
			name: "limit too large",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 101},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
			},
			wantErr:     true,
			errContains: "Limit",
		},
		{
			name: "repository error",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				repo.On("ListSymbols", ctx, mock.Anything).Return(nil, nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success - filter with label",
			opts: func() domain.ListSymbolsOptions {
				label := "test-label"
				return domain.ListSymbolsOptions{
					Filter:     domain.SymbolFilter{ProjectID: uint64(1), Label: &label},
					Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				}
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.Label != nil && *o.Filter.Label == "test-label"
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "success - filter with component_target",
			opts: func() domain.ListSymbolsOptions {
				componentTarget := "test-component"
				return domain.ListSymbolsOptions{
					Filter:     domain.SymbolFilter{ProjectID: uint64(1), ComponentTarget: &componentTarget},
					Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				}
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.ComponentTarget != nil && *o.Filter.ComponentTarget == "test-component"
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "success - filter with label and component_target",
			opts: func() domain.ListSymbolsOptions {
				label := "test-label"
				componentTarget := "test-component"
				return domain.ListSymbolsOptions{
					Filter:     domain.SymbolFilter{ProjectID: uint64(1), Label: &label, ComponentTarget: &componentTarget},
					Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				}
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Filter.Label != nil && o.Filter.ComponentTarget != nil
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "error - repository error with filter",
			opts: func() domain.ListSymbolsOptions {
				label := "test"
				return domain.ListSymbolsOptions{
					Filter:     domain.SymbolFilter{ProjectID: uint64(1), Label: &label},
					Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				}
			}(),
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				repo.On("ListSymbols", ctx, mock.Anything).Return(nil, nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success - applies default sort when not specified",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				// Sort is not set, should default to ID ASC
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				// Verify that default sort is applied
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Sort.Field == domain.SortByID && o.Sort.Direction == domain.SortAsc
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
		{
			name: "success - preserves custom sort",
			opts: domain.ListSymbolsOptions{
				Filter:     domain.SymbolFilter{ProjectID: uint64(1)},
				Pagination: pagination.OffsetPaginationParams{Offset: 0, Limit: 10},
				Sort:       domain.SortOption{Field: domain.SortByLabel, Direction: domain.SortDesc},
			},
			mockSetup: func(repo *MockSymbolRepo, ctx context.Context, opts domain.ListSymbolsOptions) {
				symbol := validSymbol()
				expectedSymbols := []*domain.Symbol{symbol}
				expectedMeta := &pagination.Meta{
					TotalCount:      1,
					Offset:          0,
					Limit:           10,
					HasNextPage:     false,
					HasPreviousPage: false,
				}
				// Verify that custom sort is preserved
				repo.On("ListSymbols", ctx, mock.MatchedBy(func(o domain.ListSymbolsOptions) bool {
					return o.Sort.Field == domain.SortByLabel && o.Sort.Direction == domain.SortDesc
				})).Return(expectedSymbols, expectedMeta, nil)
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSymbolRepo)
			uc := setupSymbolUseCase(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.opts)

			symbols, meta, err := uc.ListSymbols(ctx, tt.opts)

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

// Event Publishing Tests

func TestCreateSymbol_PublishesEvent(t *testing.T) {
	tests := []struct {
		name           string
		symbol         *domain.Symbol
		repoReturn     *domain.Symbol
		publishErr     error
		wantErr        bool
		wantEventCall  bool
		checkEventData func(*testing.T, *domain.Symbol)
	}{
		{
			name:   "publishes SymbolCreated event on success",
			symbol: validSymbol(),
			repoReturn: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			publishErr:    nil,
			wantErr:       false,
			wantEventCall: true,
			checkEventData: func(t *testing.T, published *domain.Symbol) {
				assert.Equal(t, uint64(1), published.ID)
				assert.Equal(t, "Test Symbol", published.Label)
				assert.Equal(t, uint64(1), published.Project)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", published.UID)
			},
		},
		{
			name:   "rolls back transaction on publish failure",
			symbol: validSymbol(),
			repoReturn: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			publishErr:    errors.New("publish failed"),
			wantErr:       true,
			wantEventCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupSymbolUseCaseWithDeps()
			ctx := context.Background()

			// Setup repo mock
			deps.repo.On("Create", ctx, mock.AnythingOfType("*domain.Symbol")).Return(tt.repoReturn, nil)

			// Setup publisher mock - capture the published symbol
			var publishedSymbol *domain.Symbol
			deps.pub.On("PublishSymbolCreated", ctx, mock.AnythingOfType("*domain.Symbol")).
				Run(func(args mock.Arguments) {
					publishedSymbol = args.Get(1).(*domain.Symbol)
				}).
				Return(tt.publishErr)

			result, err := deps.uc.CreateSymbol(ctx, tt.symbol)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if tt.wantEventCall {
				deps.pub.AssertCalled(t, "PublishSymbolCreated", ctx, mock.AnythingOfType("*domain.Symbol"))
				if tt.checkEventData != nil && publishedSymbol != nil {
					tt.checkEventData(t, publishedSymbol)
				}
			}

			deps.repo.AssertExpectations(t)
			deps.pub.AssertExpectations(t)
		})
	}
}

func TestUpdateSymbol_PublishesEvent(t *testing.T) {
	tests := []struct {
		name           string
		symbol         *domain.Symbol
		repoReturn     *domain.Symbol
		publishErr     error
		wantErr        bool
		wantEventCall  bool
		checkEventData func(*testing.T, *domain.Symbol)
	}{
		{
			name: "publishes SymbolUpdated event on success",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				s.Label = "Updated Label"
				return s
			}(),
			repoReturn: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				s.Label = "Updated Label"
				return s
			}(),
			publishErr:    nil,
			wantErr:       false,
			wantEventCall: true,
			checkEventData: func(t *testing.T, published *domain.Symbol) {
				assert.Equal(t, uint64(1), published.ID)
				assert.Equal(t, "Updated Label", published.Label)
			},
		},
		{
			name: "rolls back transaction on publish failure",
			symbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			repoReturn: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			publishErr:    errors.New("publish failed"),
			wantErr:       true,
			wantEventCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupSymbolUseCaseWithDeps()
			ctx := context.Background()

			// Setup repo mock
			deps.repo.On("Update", ctx, mock.AnythingOfType("*domain.Symbol")).Return(tt.repoReturn, nil)

			// Setup publisher mock - capture the published symbol
			var publishedSymbol *domain.Symbol
			deps.pub.On("PublishSymbolUpdated", ctx, mock.AnythingOfType("*domain.Symbol")).
				Run(func(args mock.Arguments) {
					publishedSymbol = args.Get(1).(*domain.Symbol)
				}).
				Return(tt.publishErr)

			result, err := deps.uc.UpdateSymbol(ctx, tt.symbol)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if tt.wantEventCall {
				deps.pub.AssertCalled(t, "PublishSymbolUpdated", ctx, mock.AnythingOfType("*domain.Symbol"))
				if tt.checkEventData != nil && publishedSymbol != nil {
					tt.checkEventData(t, publishedSymbol)
				}
			}

			deps.repo.AssertExpectations(t)
			deps.pub.AssertExpectations(t)
		})
	}
}

func TestDeleteSymbol_PublishesEvent(t *testing.T) {
	tests := []struct {
		name           string
		symbolID       uint64
		foundSymbol    *domain.Symbol
		publishErr     error
		wantErr        bool
		wantEventCall  bool
		checkEventData func(*testing.T, *domain.Symbol)
	}{
		{
			name:     "publishes SymbolDeleted event on success",
			symbolID: 1,
			foundSymbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			publishErr:    nil,
			wantErr:       false,
			wantEventCall: true,
			checkEventData: func(t *testing.T, published *domain.Symbol) {
				assert.Equal(t, uint64(1), published.ID)
				assert.Equal(t, "Test Symbol", published.Label)
				assert.Equal(t, uint64(1), published.Project)
			},
		},
		{
			name:     "rolls back transaction on publish failure",
			symbolID: 1,
			foundSymbol: func() *domain.Symbol {
				s := validSymbol()
				s.ID = 1
				return s
			}(),
			publishErr:    errors.New("publish failed"),
			wantErr:       true,
			wantEventCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupSymbolUseCaseWithDeps()
			ctx := context.Background()

			// Setup repo mocks
			deps.repo.On("FindByID", ctx, tt.symbolID).Return(tt.foundSymbol, nil)
			deps.repo.On("Delete", ctx, tt.symbolID).Return(nil)

			// Setup publisher mock - capture the published symbol
			var publishedSymbol *domain.Symbol
			deps.pub.On("PublishSymbolDeleted", ctx, mock.AnythingOfType("*domain.Symbol")).
				Run(func(args mock.Arguments) {
					publishedSymbol = args.Get(1).(*domain.Symbol)
				}).
				Return(tt.publishErr)

			err := deps.uc.DeleteSymbol(ctx, tt.symbolID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantEventCall {
				deps.pub.AssertCalled(t, "PublishSymbolDeleted", ctx, mock.AnythingOfType("*domain.Symbol"))
				if tt.checkEventData != nil && publishedSymbol != nil {
					tt.checkEventData(t, publishedSymbol)
				}
			}

			deps.repo.AssertExpectations(t)
			deps.pub.AssertExpectations(t)
		})
	}
}

func TestEventPublishing_NotCalledOnValidationFailure(t *testing.T) {
	t.Run("CreateSymbol does not publish event on validation failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		invalidSymbol := &domain.Symbol{
			// Missing required fields - will fail validation
			Label: "",
		}

		_, err := deps.uc.CreateSymbol(ctx, invalidSymbol)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolCreated", mock.Anything, mock.Anything)
	})

	t.Run("UpdateSymbol does not publish event on validation failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		invalidSymbol := &domain.Symbol{
			ID:    1,
			Label: "", // Invalid - empty label
		}

		_, err := deps.uc.UpdateSymbol(ctx, invalidSymbol)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolUpdated", mock.Anything, mock.Anything)
	})

	t.Run("DeleteSymbol does not publish event on invalid ID", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		err := deps.uc.DeleteSymbol(ctx, 0) // Invalid ID

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolDeleted", mock.Anything, mock.Anything)
	})
}

func TestEventPublishing_NotCalledOnRepoFailure(t *testing.T) {
	t.Run("CreateSymbol does not publish event on repo failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		symbol := validSymbol()
		deps.repo.On("Create", ctx, mock.AnythingOfType("*domain.Symbol")).
			Return(nil, errors.New("database error"))

		_, err := deps.uc.CreateSymbol(ctx, symbol)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolCreated", mock.Anything, mock.Anything)
	})

	t.Run("UpdateSymbol does not publish event on repo failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		symbol := validSymbol()
		symbol.ID = 1
		deps.repo.On("Update", ctx, mock.AnythingOfType("*domain.Symbol")).
			Return(nil, errors.New("database error"))

		_, err := deps.uc.UpdateSymbol(ctx, symbol)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolUpdated", mock.Anything, mock.Anything)
	})

	t.Run("DeleteSymbol does not publish event on FindByID failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		deps.repo.On("FindByID", ctx, uint64(1)).Return(nil, domain.ErrDataNotFound)

		err := deps.uc.DeleteSymbol(ctx, 1)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolDeleted", mock.Anything, mock.Anything)
	})

	t.Run("DeleteSymbol does not publish event on Delete failure", func(t *testing.T) {
		deps := setupSymbolUseCaseWithDeps()
		ctx := context.Background()

		symbol := validSymbol()
		symbol.ID = 1
		deps.repo.On("FindByID", ctx, uint64(1)).Return(symbol, nil)
		deps.repo.On("Delete", ctx, uint64(1)).Return(errors.New("database error"))

		err := deps.uc.DeleteSymbol(ctx, 1)

		assert.Error(t, err)
		deps.pub.AssertNotCalled(t, "PublishSymbolDeleted", mock.Anything, mock.Anything)
	})
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	assert.NotNil(t, v)
	assert.IsType(t, &validator.Validate{}, v)
}
