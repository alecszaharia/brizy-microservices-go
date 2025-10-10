package biz

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSymbolRepo is a mock implementation of SymbolRepo for testing
type MockSymbolRepo struct {
	mock.Mock
}

func (m *MockSymbolRepo) Save(ctx context.Context, symbol *Symbol) (*Symbol, error) {
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

func (m *MockSymbolRepo) FindByID(ctx context.Context, id int32) (*Symbol, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Symbol), args.Error(1)
}

func (m *MockSymbolRepo) ListSymbols(ctx context.Context, options *ListSymbolsOptions) ([]*Symbol, *SymbolCursor, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*Symbol), args.Get(1).(*SymbolCursor), args.Error(2)
}

func (m *MockSymbolRepo) Delete(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Helper function to create a test SymbolUseCase
func setupSymbolUseCase(mockRepo *MockSymbolRepo) *SymbolUseCase {
	logger := log.NewStdLogger(os.Stdout)
	v := NewSymbolValidator()
	return NewSymbolUseCase(mockRepo, v, logger)
}

// Helper function to create a valid Symbol for testing
func validSymbol() *Symbol {
	return &Symbol{
		Project:         1,
		Uid:             "550e8400-e29b-41d4-a716-446655440000",
		Label:           "Test Symbol",
		ClassName:       "TestClass",
		ComponentTarget: "TestTarget",
		Version:         "1.0.0",
		Data:            &SymbolData{Project: 1, Data: []byte(`{"key": "value"}`)},
	}
}

func TestCreateSymbol_Success(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	expectedSymbol := &Symbol{
		Id:              1,
		Project:         symbol.Project,
		Uid:             symbol.Uid,
		Label:           symbol.Label,
		ClassName:       symbol.ClassName,
		ComponentTarget: symbol.ComponentTarget,
		Version:         symbol.Version,
		Data:            symbol.Data,
	}

	mockRepo.On("Save", ctx, symbol).Return(expectedSymbol, nil)

	result, err := uc.CreateSymbol(ctx, symbol)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.Id)
	assert.Equal(t, symbol.Label, result.Label)
	mockRepo.AssertExpectations(t)
}

func TestCreateSymbol_ValidationErrors(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		symbol        *Symbol
		expectedError string
	}{
		{
			name: "Missing Project",
			symbol: &Symbol{
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         "1.0.0",
			},
			expectedError: "Project",
		},
		{
			name: "Invalid UUID",
			symbol: &Symbol{
				Project:         1,
				Uid:             "invalid-uuid",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         "1.0.0",
			},
			expectedError: "Uid",
		},
		{
			name: "Empty Label",
			symbol: &Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         "1.0.0",
			},
			expectedError: "Label",
		},
		{
			name: "Empty ClassName",
			symbol: &Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "",
				ComponentTarget: "Target",
				Version:         "1.0.0",
			},
			expectedError: "ClassName",
		},
		{
			name: "Empty ComponentTarget",
			symbol: &Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "",
				Version:         "1.0.0",
			},
			expectedError: "ComponentTarget",
		},
		{
			name: "Invalid Version",
			symbol: &Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test",
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         "invalid",
			},
			expectedError: "Version",
		},
		{
			name: "Label Too Long",
			symbol: &Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           string(make([]byte, 256)),
				ClassName:       "Class",
				ComponentTarget: "Target",
				Version:         "1.0.0",
			},
			expectedError: "Label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uc.CreateSymbol(ctx, tt.symbol)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestCreateSymbol_RepoError(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	expectedError := errors.New("database error")

	mockRepo.On("Save", ctx, symbol).Return(nil, expectedError)

	result, err := uc.CreateSymbol(ctx, symbol)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSymbol_Success(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	symbol.Id = 1

	mockRepo.On("Update", ctx, symbol).Return(symbol, nil)

	result, err := uc.UpdateSymbol(ctx, symbol)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, symbol.Id, result.Id)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSymbol_ValidationError(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	symbol.Id = 1
	symbol.Label = "" // Invalid

	result, err := uc.UpdateSymbol(ctx, symbol)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateSymbol_RepoError(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	symbol.Id = 1
	expectedError := errors.New("update failed")

	mockRepo.On("Update", ctx, symbol).Return(nil, expectedError)

	result, err := uc.UpdateSymbol(ctx, symbol)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteSymbol_Success(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbolID := int32(1)

	mockRepo.On("Delete", ctx, symbolID).Return(nil)

	err := uc.DeleteSymbol(ctx, symbolID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteSymbol_InvalidID(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name string
		id   int32
	}{
		{"Zero ID", 0},
		{"Negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.DeleteSymbol(ctx, tt.id)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "INVALID_ID")
		})
	}
}

func TestDeleteSymbol_RepoError(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbolID := int32(1)
	expectedError := errors.New("delete failed")

	mockRepo.On("Delete", ctx, symbolID).Return(expectedError)

	err := uc.DeleteSymbol(ctx, symbolID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestListSymbols_Success(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	options := &ListSymbolsOptions{
		ProjectID: 1,
		PageSize:  10,
	}

	expectedSymbols := []*Symbol{
		validSymbol(),
	}
	expectedCursor := &SymbolCursor{
		LastID:    1,
		OrderBy:   "id",
		Direction: "ASC",
	}

	mockRepo.On("ListSymbols", ctx, options).Return(expectedSymbols, expectedCursor, nil)

	symbols, cursor, err := uc.ListSymbols(ctx, options)

	assert.NoError(t, err)
	assert.NotNil(t, symbols)
	assert.NotNil(t, cursor)
	assert.Len(t, symbols, 1)
	assert.Equal(t, expectedCursor.LastID, cursor.LastID)
	mockRepo.AssertExpectations(t)
}

func TestListSymbols_WithCursor(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	options := &ListSymbolsOptions{
		ProjectID: 1,
		PageSize:  10,
		Cursor: &SymbolCursor{
			LastID:    5,
			OrderBy:   "label",
			Direction: "DESC",
		},
	}

	expectedSymbols := []*Symbol{validSymbol()}
	expectedCursor := &SymbolCursor{
		LastID:    10,
		OrderBy:   "label",
		Direction: "DESC",
	}

	mockRepo.On("ListSymbols", ctx, options).Return(expectedSymbols, expectedCursor, nil)

	symbols, cursor, err := uc.ListSymbols(ctx, options)

	assert.NoError(t, err)
	assert.NotNil(t, symbols)
	assert.NotNil(t, cursor)
	mockRepo.AssertExpectations(t)
}

func TestListSymbols_ValidationErrors(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		options       *ListSymbolsOptions
		expectedError string
	}{
		{
			name: "Missing ProjectID",
			options: &ListSymbolsOptions{
				PageSize: 10,
			},
			expectedError: "ProjectID",
		},
		{
			name: "Zero ProjectID",
			options: &ListSymbolsOptions{
				ProjectID: 0,
				PageSize:  10,
			},
			expectedError: "ProjectID",
		},
		{
			name: "Missing PageSize",
			options: &ListSymbolsOptions{
				ProjectID: 1,
			},
			expectedError: "PageSize",
		},
		{
			name: "PageSize Too Large",
			options: &ListSymbolsOptions{
				ProjectID: 1,
				PageSize:  101,
			},
			expectedError: "PageSize",
		},
		{
			name: "Invalid Cursor OrderBy",
			options: &ListSymbolsOptions{
				ProjectID: 1,
				PageSize:  10,
				Cursor: &SymbolCursor{
					LastID:    1,
					OrderBy:   "invalid",
					Direction: "ASC",
				},
			},
			expectedError: "OrderBy",
		},
		{
			name: "Invalid Cursor Direction",
			options: &ListSymbolsOptions{
				ProjectID: 1,
				PageSize:  10,
				Cursor: &SymbolCursor{
					LastID:    1,
					OrderBy:   "id",
					Direction: "INVALID",
				},
			},
			expectedError: "Direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbols, cursor, err := uc.ListSymbols(ctx, tt.options)

			assert.Error(t, err)
			assert.Nil(t, symbols)
			assert.Nil(t, cursor)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestListSymbols_RepoError(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	options := &ListSymbolsOptions{
		ProjectID: 1,
		PageSize:  10,
	}
	expectedError := errors.New("database error")

	mockRepo.On("ListSymbols", ctx, options).Return(nil, nil, expectedError)

	symbols, cursor, err := uc.ListSymbols(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, symbols)
	assert.Nil(t, cursor)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestSymbol_WithSymbolData(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	symbol.Data = &SymbolData{
		Project: 1,
		Data:    []byte(`{"key": "value"}`),
	}

	expectedSymbol := &Symbol{
		Id:              1,
		Project:         symbol.Project,
		Uid:             symbol.Uid,
		Label:           symbol.Label,
		ClassName:       symbol.ClassName,
		ComponentTarget: symbol.ComponentTarget,
		Version:         symbol.Version,
		Data:            symbol.Data,
	}

	mockRepo.On("Save", ctx, symbol).Return(expectedSymbol, nil)

	result, err := uc.CreateSymbol(ctx, symbol)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)
	assert.Equal(t, symbol.Data.Data, result.Data.Data)
	mockRepo.AssertExpectations(t)
}

func TestSymbol_SymbolDataValidation(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	uc := setupSymbolUseCase(mockRepo)
	ctx := context.Background()

	symbol := validSymbol()
	symbol.Data = &SymbolData{
		Project: 0, // Invalid: must be gt=0
		Data:    []byte(`{"key": "value"}`),
	}

	result, err := uc.CreateSymbol(ctx, symbol)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Project")
}

func TestNewSymbolValidator(t *testing.T) {
	v := NewSymbolValidator()
	assert.NotNil(t, v)
	assert.IsType(t, &validator.Validate{}, v)
}

func TestNewSymbolUseCase(t *testing.T) {
	mockRepo := new(MockSymbolRepo)
	logger := log.NewStdLogger(nil)
	v := NewSymbolValidator()

	uc := NewSymbolUseCase(mockRepo, v, logger)

	assert.NotNil(t, uc)
	assert.Equal(t, mockRepo, uc.repo)
	assert.NotNil(t, uc.log)
	assert.Equal(t, v, uc.validator)
}
