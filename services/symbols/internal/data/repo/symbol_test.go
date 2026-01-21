package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"platform/pagination"
	"symbols/internal/biz/domain"
	"symbols/internal/data/common"
	"symbols/internal/data/model"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to open in-memory SQLite database: %v", err)
	}

	// Run migrations for test tables
	if err := db.AutoMigrate(&model.Symbol{}, &model.SymbolData{}); err != nil {
		t.Errorf("Failed to migrate test tables: %v", err)
	}

	return db
}

// cleanupDB cleans up test data
func cleanupDB(db *gorm.DB) {
	db.Exec("DELETE FROM symbol_data")
	db.Exec("DELETE FROM symbols")
}

// validDomainSymbol returns a valid domain symbol for testing
func validDomainSymbol() *domain.Symbol {
	data := []byte(`{"key": "value"}`)
	return &domain.Symbol{
		Project:         1,
		UID:             "550e8400-e29b-41d4-a716-446655440000",
		Label:           "Test Symbol",
		ClassName:       "TestClass",
		ComponentTarget: "TestTarget",
		Version:         1,
		Data: &domain.SymbolData{
			Project: 1,
			Data:    &data,
		},
	}
}

// validEntitySymbol returns a valid GORM entity for testing
func validEntitySymbol() *model.Symbol {
	data := []byte(`{"key": "value"}`)
	return &model.Symbol{
		ProjectID:       1,
		UID:             "550e8400-e29b-41d4-a716-446655440000",
		Label:           "Test Symbol",
		ClassName:       "TestClass",
		ComponentTarget: "TestTarget",
		Version:         1,
		SymbolData: &model.SymbolData{
			Data: &data,
		},
	}
}

// mockTransaction is a mock implementation of the Transaction interface
type mockTransaction struct{}

func (m *mockTransaction) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Compile-time interface check
var _ common.Transaction = (*mockTransaction)(nil)

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		input       *domain.Symbol
		setup       func(*gorm.DB)
		wantErr     bool
		checkError  func(*testing.T, error)
		checkResult func(*testing.T, *domain.Symbol)
	}{
		{
			name:    "success with nested data",
			input:   validDomainSymbol(),
			setup:   func(db *gorm.DB) {},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotZero(t, result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
				assert.NotNil(t, result.Data)
				assert.NotZero(t, result.Data.ID)
			},
		},
		{
			name: "success without nested data",
			input: func() *domain.Symbol {
				s := validDomainSymbol()
				s.Data = nil
				return s
			}(),
			setup:   func(db *gorm.DB) {},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotZero(t, result.ID)
				assert.Nil(t, result.Data)
			},
		},
		{
			name:  "duplicate entry error",
			input: validDomainSymbol(),
			setup: func(db *gorm.DB) {
				entity := validEntitySymbol()
				db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrDataDuplicateEntry)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer cleanupDB(db)

			tt.setup(db)

			logger := log.NewStdLogger(os.Stdout)
			tx := &mockTransaction{}
			repo := NewSymbolRepo(db, tx, logger)

			result, err := repo.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		input       *domain.Symbol
		setup       func(*gorm.DB) uint64
		wantErr     bool
		checkError  func(*testing.T, error)
		checkResult func(*testing.T, *domain.Symbol)
	}{
		{
			name:  "success",
			input: validDomainSymbol(),
			setup: func(db *gorm.DB) uint64 {
				entity := validEntitySymbol()
				db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				return entity.ID
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotZero(t, result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
			},
		},
		{
			name: "success - update label",
			input: func() *domain.Symbol {
				s := validDomainSymbol()
				s.Label = "Updated Label"
				return s
			}(),
			setup: func(db *gorm.DB) uint64 {
				entity := validEntitySymbol()
				db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				return entity.ID
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, "Updated Label", result.Label)
			},
		},
		{
			name: "not found error",
			input: func() *domain.Symbol {
				s := validDomainSymbol()
				s.ID = 999
				return s
			}(),
			setup: func(db *gorm.DB) uint64 {
				return 999
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrDataNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer cleanupDB(db)

			id := tt.setup(db)
			tt.input.ID = id

			logger := log.NewStdLogger(os.Stdout)
			tx := &mockTransaction{}
			repo := NewSymbolRepo(db, tx, logger)

			result, err := repo.Update(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestFindByID(t *testing.T) {
	tests := []struct {
		name        string
		id          uint64
		setup       func(*gorm.DB) uint64
		wantErr     bool
		checkError  func(*testing.T, error)
		checkResult func(*testing.T, *domain.Symbol)
	}{
		{
			name: "success with nested data",
			setup: func(db *gorm.DB) uint64 {
				entity := validEntitySymbol()
				db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				return entity.ID
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotZero(t, result.ID)
				assert.Equal(t, "Test Symbol", result.Label)
				assert.NotNil(t, result.Data)
			},
		},
		{
			name: "success without nested data",
			setup: func(db *gorm.DB) uint64 {
				entity := validEntitySymbol()
				entity.SymbolData = nil
				db.Create(entity)
				return entity.ID
			},
			wantErr: false,
			checkResult: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.NotZero(t, result.ID)
				assert.Nil(t, result.Data)
			},
		},
		{
			name: "not found error",
			setup: func(db *gorm.DB) uint64 {
				return 999
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrDataNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer cleanupDB(db)

			id := tt.setup(db)

			logger := log.NewStdLogger(os.Stdout)
			tx := &mockTransaction{}
			repo := NewSymbolRepo(db, tx, logger)

			result, err := repo.FindByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestListSymbols(t *testing.T) {
	tests := []struct {
		name        string
		offset      uint64
		limit       uint32
		filter      map[string]interface{}
		setup       func(*gorm.DB)
		wantErr     bool
		checkResult func(*testing.T, []*domain.Symbol, *pagination.Meta)
	}{
		{
			name:   "success - first page with next page",
			offset: 0,
			limit:  5,
			filter: map[string]interface{}{"project_id": uint64(1)},
			setup: func(db *gorm.DB) {
				for i := 0; i < 10; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 5)
				assert.Equal(t, uint64(10), meta.TotalCount)
				assert.True(t, meta.HasNextPage)
				assert.False(t, meta.HasPreviousPage)
			},
		},
		{
			name:   "success - second page",
			offset: 5,
			limit:  5,
			filter: map[string]interface{}{"project_id": uint64(1)},
			setup: func(db *gorm.DB) {
				for i := 0; i < 10; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 5)
				assert.Equal(t, uint64(10), meta.TotalCount)
				assert.False(t, meta.HasNextPage)
				assert.True(t, meta.HasPreviousPage)
			},
		},
		{
			name:    "success - empty results",
			offset:  0,
			limit:   10,
			filter:  map[string]interface{}{"project_id": uint64(999)},
			setup:   func(db *gorm.DB) {},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Empty(t, symbols)
				assert.Equal(t, uint64(0), meta.TotalCount)
				assert.False(t, meta.HasNextPage)
				assert.False(t, meta.HasPreviousPage)
			},
		},
		{
			name:   "success - last page incomplete",
			offset: 5,
			limit:  10,
			filter: map[string]interface{}{"project_id": uint64(1)},
			setup: func(db *gorm.DB) {
				for i := 0; i < 8; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 3)
				assert.Equal(t, uint64(8), meta.TotalCount)
				assert.False(t, meta.HasNextPage)
				assert.True(t, meta.HasPreviousPage)
			},
		},
		{
			name:   "success - filter by multiple fields",
			offset: 0,
			limit:  10,
			filter: map[string]interface{}{"project_id": uint64(1), "label": "test-label"},
			setup: func(db *gorm.DB) {
				// Create symbols with matching label
				for i := 0; i < 3; i++ {
					entity := validEntitySymbol()
					entity.Label = "test-label"
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
				// Create symbols with different label
				for i := 3; i < 7; i++ {
					entity := validEntitySymbol()
					entity.Label = "other-label"
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 3)
				assert.Equal(t, uint64(3), meta.TotalCount)
				for _, s := range symbols {
					assert.Equal(t, "test-label", s.Label)
				}
			},
		},
		{
			name:   "success - nil filter returns all",
			offset: 0,
			limit:  10,
			filter: nil,
			setup: func(db *gorm.DB) {
				for i := 0; i < 5; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 5)
				assert.Equal(t, uint64(5), meta.TotalCount)
			},
		},
		{
			name:   "success - empty filter returns all",
			offset: 0,
			limit:  10,
			filter: map[string]interface{}{},
			setup: func(db *gorm.DB) {
				for i := 0; i < 5; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 5)
				assert.Equal(t, uint64(5), meta.TotalCount)
			},
		},
		{
			name:   "success - filter with pagination interaction",
			offset: 2,
			limit:  3,
			filter: map[string]interface{}{"project_id": uint64(1)},
			setup: func(db *gorm.DB) {
				// Create 10 symbols for project 1
				for i := 0; i < 10; i++ {
					entity := validEntitySymbol()
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
				// Create 5 symbols for project 2 (should be filtered out)
				for i := 10; i < 15; i++ {
					entity := validEntitySymbol()
					entity.ProjectID = 2
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Len(t, symbols, 3)
				// Total count should reflect only filtered results (10), not all records (15)
				assert.Equal(t, uint64(10), meta.TotalCount)
				assert.True(t, meta.HasNextPage)
				assert.True(t, meta.HasPreviousPage)
			},
		},
		{
			name:   "success - zero results with filter",
			offset: 0,
			limit:  10,
			filter: map[string]interface{}{"label": "non-existent-label"},
			setup: func(db *gorm.DB) {
				for i := 0; i < 5; i++ {
					entity := validEntitySymbol()
					entity.Label = "existing-label"
					entity.UID = fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", i)
					db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, symbols []*domain.Symbol, meta *pagination.Meta) {
				assert.Empty(t, symbols)
				assert.Equal(t, uint64(0), meta.TotalCount)
				assert.False(t, meta.HasNextPage)
				assert.False(t, meta.HasPreviousPage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer cleanupDB(db)

			tt.setup(db)

			logger := log.NewStdLogger(os.Stdout)
			tx := &mockTransaction{}
			repo := NewSymbolRepo(db, tx, logger)

			symbols, meta, err := repo.ListSymbols(context.Background(), tt.offset, tt.limit, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, symbols)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, symbols, meta)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		id         uint64
		setup      func(*gorm.DB) uint64
		wantErr    bool
		checkError func(*testing.T, error)
	}{
		{
			name: "success",
			setup: func(db *gorm.DB) uint64 {
				entity := validEntitySymbol()
				db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity)
				return entity.ID
			},
			wantErr: false,
		},
		{
			name: "not found error",
			setup: func(db *gorm.DB) uint64 {
				return 999
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrDataNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer cleanupDB(db)

			id := tt.setup(db)

			logger := log.NewStdLogger(os.Stdout)
			tx := &mockTransaction{}
			repo := NewSymbolRepo(db, tx, logger)

			err := repo.Delete(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)

				// Verify symbol was soft deleted
				var count int64
				db.Unscoped().Model(&model.Symbol{}).Where("id = ?", id).Count(&count)
				assert.Equal(t, int64(1), count)

				// Verify it's not visible in normal queries
				var normalCount int64
				db.Model(&model.Symbol{}).Where("id = ?", id).Count(&normalCount)
				assert.Equal(t, int64(0), normalCount)
			}
		})
	}
}

func TestToDomainSymbol(t *testing.T) {
	tests := []struct {
		name   string
		input  *model.Symbol
		assert func(*testing.T, *domain.Symbol)
	}{
		{
			name:  "nil input returns nil",
			input: nil,
			assert: func(t *testing.T, result *domain.Symbol) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "complete entity converts correctly",
			input: validEntitySymbol(),
			assert: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(0), result.ID) // Entity ID not set
				assert.Equal(t, uint64(1), result.Project)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.UID)
				assert.Equal(t, "Test Symbol", result.Label)
				assert.Equal(t, "TestClass", result.ClassName)
				assert.Equal(t, "TestTarget", result.ComponentTarget)
				assert.Equal(t, uint32(1), result.Version)
				assert.NotNil(t, result.Data)
				assert.Equal(t, uint64(1), result.Data.Project)
			},
		},
		{
			name: "entity without nested data",
			input: func() *model.Symbol {
				e := validEntitySymbol()
				e.SymbolData = nil
				return e
			}(),
			assert: func(t *testing.T, result *domain.Symbol) {
				assert.NotNil(t, result)
				assert.Nil(t, result.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDomainSymbol(tt.input)
			tt.assert(t, result)
		})
	}
}

func TestToEntitySymbol(t *testing.T) {
	tests := []struct {
		name   string
		input  *domain.Symbol
		assert func(*testing.T, *model.Symbol)
	}{
		{
			name:  "nil input returns nil",
			input: nil,
			assert: func(t *testing.T, result *model.Symbol) {
				assert.Nil(t, result)
			},
		},
		{
			name:  "complete domain converts correctly",
			input: validDomainSymbol(),
			assert: func(t *testing.T, result *model.Symbol) {
				assert.NotNil(t, result)
				assert.Equal(t, uint64(0), result.ID) // Domain ID not set
				assert.Equal(t, uint64(1), result.ProjectID)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.UID)
				assert.Equal(t, "Test Symbol", result.Label)
				assert.Equal(t, "TestClass", result.ClassName)
				assert.Equal(t, "TestTarget", result.ComponentTarget)
				assert.Equal(t, uint32(1), result.Version)
				assert.NotNil(t, result.SymbolData)
			},
		},
		{
			name: "domain without nested data",
			input: func() *domain.Symbol {
				s := validDomainSymbol()
				s.Data = nil
				return s
			}(),
			assert: func(t *testing.T, result *model.Symbol) {
				assert.NotNil(t, result)
				assert.Nil(t, result.SymbolData)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toEntitySymbol(tt.input)
			tt.assert(t, result)
		})
	}
}

func TestMapGormError(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(db)

	logger := log.NewStdLogger(os.Stdout)
	tx := &mockTransaction{}
	repo := NewSymbolRepo(db, tx, logger).(*symbolRepo)

	tests := []struct {
		name       string
		inputError error
		wantError  error
	}{
		{
			name:       "nil input returns nil",
			inputError: nil,
			wantError:  nil,
		},
		{
			name:       "gorm.ErrRecordNotFound returns domain.ErrDataNotFound",
			inputError: gorm.ErrRecordNotFound,
			wantError:  domain.ErrDataNotFound,
		},
		{
			name:       "MySQL duplicate key error",
			inputError: errors.New("Error 1062: Duplicate entry '1-uid' for key 'idx_project_uid'"),
			wantError:  domain.ErrDataDuplicateEntry,
		},
		{
			name:       "duplicate entry message",
			inputError: errors.New("Duplicate entry '100-uid' for key 'idx_project_uid'"),
			wantError:  domain.ErrDataDuplicateEntry,
		},
		{
			name:       "SQLite unique constraint",
			inputError: errors.New("UNIQUE constraint failed: symbols.project_id, symbols.uid"),
			wantError:  domain.ErrDataDuplicateEntry,
		},
		{
			name:       "other errors return domain.ErrDataDatabase",
			inputError: errors.New("connection timeout"),
			wantError:  domain.ErrDataDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.mapGormError(tt.inputError)
			if tt.wantError == nil {
				assert.Nil(t, result)
			} else {
				assert.ErrorIs(t, result, tt.wantError)
			}
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupDB(db)

	logger := log.NewStdLogger(os.Stdout)
	tx := &mockTransaction{}
	repo := NewSymbolRepo(db, tx, logger).(*symbolRepo)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "MySQL Error 1062",
			err:      errors.New("Error 1062: Duplicate entry"),
			expected: true,
		},
		{
			name:     "Duplicate entry message",
			err:      errors.New("Duplicate entry 'value' for key 'idx'"),
			expected: true,
		},
		{
			name:     "SQLite unique constraint",
			err:      errors.New("UNIQUE constraint failed: table.column"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.isDuplicateKeyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
