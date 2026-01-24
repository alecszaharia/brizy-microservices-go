package service

import (
	v1 "contracts/gen/service/symbols/v1"
	"platform/pagination"
	"testing"

	"symbols/internal/biz/domain"

	"github.com/stretchr/testify/assert"
)

func Test_toBizSymbol(t *testing.T) {
	data := []byte(`{"key": "value"}`)

	tests := []struct {
		name     string
		input    *v1.Symbol
		expected *domain.Symbol
	}{
		{
			name: "complete symbol conversion",
			input: &v1.Symbol{
				Id:              123,
				ProjectId:       456,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "web",
				Version:         1,
				Data:            data,
			},
			expected: &domain.Symbol{
				ID:              123,
				Project:         456,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "web",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 456,
					Data:    &data,
				},
			},
		},
		{
			name: "symbol with empty data",
			input: &v1.Symbol{
				Id:              1,
				ProjectId:       2,
				Uid:             "550e8400-e29b-41d4-a716-446655440001",
				Label:           "Empty Data Symbol",
				ClassName:       "EmptyClass",
				ComponentTarget: "mobile",
				Version:         1,
				Data:            []byte{},
			},
			expected: &domain.Symbol{
				ID:              1,
				Project:         2,
				UID:             "550e8400-e29b-41d4-a716-446655440001",
				Label:           "Empty Data Symbol",
				ClassName:       "EmptyClass",
				ComponentTarget: "mobile",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 2,
					Data:    &[]byte{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBizSymbol(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Project, result.Project)
			assert.Equal(t, tt.expected.UID, result.UID)
			assert.Equal(t, tt.expected.Label, result.Label)
			assert.Equal(t, tt.expected.ClassName, result.ClassName)
			assert.Equal(t, tt.expected.ComponentTarget, result.ComponentTarget)
			assert.Equal(t, tt.expected.Version, result.Version)

			assert.NotNil(t, result.Data)
			assert.Equal(t, tt.expected.Data.Project, result.Data.Project)
			assert.Equal(t, *tt.expected.Data.Data, *result.Data.Data)
		})
	}
}

func Test_toBizSymbolFromRequest(t *testing.T) {
	data := []byte(`{"config": "test"}`)

	tests := []struct {
		name     string
		input    *v1.CreateSymbolRequest
		expected *domain.Symbol
	}{
		{
			name: "complete create request conversion",
			input: &v1.CreateSymbolRequest{
				ProjectId:       789,
				Uid:             "550e8400-e29b-41d4-a716-446655440002",
				Label:           "New Symbol",
				ClassName:       "NewClass",
				ComponentTarget: "desktop",
				Version:         1,
				Data:            data,
			},
			expected: &domain.Symbol{
				Project:         789,
				UID:             "550e8400-e29b-41d4-a716-446655440002",
				Label:           "New Symbol",
				ClassName:       "NewClass",
				ComponentTarget: "desktop",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 789,
					Data:    &data,
				},
			},
		},
		{
			name: "create request with empty data",
			input: &v1.CreateSymbolRequest{
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440003",
				Label:           "Min Symbol",
				ClassName:       "MinClass",
				ComponentTarget: "app",
				Version:         1,
				Data:            []byte{},
			},
			expected: &domain.Symbol{
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440003",
				Label:           "Min Symbol",
				ClassName:       "MinClass",
				ComponentTarget: "app",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 1,
					Data:    &[]byte{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SymbolFromCreateRequest(tt.input)

			assert.Equal(t, tt.expected.Project, result.Project)
			assert.Equal(t, tt.expected.UID, result.UID)
			assert.Equal(t, tt.expected.Label, result.Label)
			assert.Equal(t, tt.expected.ClassName, result.ClassName)
			assert.Equal(t, tt.expected.ComponentTarget, result.ComponentTarget)
			assert.Equal(t, tt.expected.Version, result.Version)

			assert.NotNil(t, result.Data)
			assert.Equal(t, tt.expected.Data.Project, result.Data.Project)
			assert.Equal(t, tt.expected.Data.Data, result.Data.Data)
		})
	}
}

func Test_toV1Symbol(t *testing.T) {
	data := []byte(`{"test": "data"}`)

	tests := []struct {
		name     string
		input    *domain.Symbol
		expected *v1.Symbol
	}{
		{
			name: "complete biz symbol to v1",
			input: &domain.Symbol{
				ID:              999,
				Project:         888,
				UID:             "550e8400-e29b-41d4-a716-446655440004",
				Label:           "Converted Symbol",
				ClassName:       "ConvertedClass",
				ComponentTarget: "universal",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 888,
					Data:    &data,
				},
			},
			expected: &v1.Symbol{
				Id:              999,
				ProjectId:       888,
				Uid:             "550e8400-e29b-41d4-a716-446655440004",
				Label:           "Converted Symbol",
				ClassName:       "ConvertedClass",
				ComponentTarget: "universal",
				Version:         1,
				Data:            data,
			},
		},
		{
			name: "biz symbol with nil data",
			input: &domain.Symbol{
				ID:              100,
				Project:         200,
				UID:             "550e8400-e29b-41d4-a716-446655440005",
				Label:           "No Data Symbol",
				ClassName:       "NoDataClass",
				ComponentTarget: "none",
				Version:         1,
				Data:            nil,
			},
			expected: &v1.Symbol{
				Id:              100,
				ProjectId:       200,
				Uid:             "550e8400-e29b-41d4-a716-446655440005",
				Label:           "No Data Symbol",
				ClassName:       "NoDataClass",
				ComponentTarget: "none",
				Version:         1,
				Data:            nil,
			},
		},
		{
			name: "biz symbol with empty data bytes",
			input: &domain.Symbol{
				ID:              101,
				Project:         201,
				UID:             "550e8400-e29b-41d4-a716-446655440006",
				Label:           "Empty Bytes Symbol",
				ClassName:       "EmptyBytesClass",
				ComponentTarget: "all",
				Version:         1,
				Data: &domain.SymbolData{
					Project: 201,
					Data:    &[]byte{},
				},
			},
			expected: &v1.Symbol{
				Id:              101,
				ProjectId:       201,
				Uid:             "550e8400-e29b-41d4-a716-446655440006",
				Label:           "Empty Bytes Symbol",
				ClassName:       "EmptyBytesClass",
				ComponentTarget: "all",
				Version:         1,
				Data:            []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toV1Symbol(tt.input)

			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.ProjectId, result.ProjectId)
			assert.Equal(t, tt.expected.Uid, result.Uid)
			assert.Equal(t, tt.expected.Label, result.Label)
			assert.Equal(t, tt.expected.ClassName, result.ClassName)
			assert.Equal(t, tt.expected.ComponentTarget, result.ComponentTarget)
			assert.Equal(t, tt.expected.Version, result.Version)
			assert.Equal(t, tt.expected.Data, result.Data)
		})
	}
}

func Test_toBizListSymbolsOptions(t *testing.T) {
	labelFilter := "test-label"
	componentTargetFilter := "test-component"

	tests := []struct {
		name     string
		input    *v1.ListSymbolsRequest
		expected domain.ListSymbolsOptions
	}{
		{
			name: "with offset and limit",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    10,
				Limit:     20,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID: 1,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 10,
					Limit:  20,
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "with default limit (zero)",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     0,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID: 1,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  20, // Default value
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "first page",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     10,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID: 1,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  10,
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "with label filter",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     10,
				Label:     &labelFilter,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID: 1,
					Label:     &labelFilter,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  10,
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "with component_target filter",
			input: &v1.ListSymbolsRequest{
				ProjectId:       1,
				Offset:          0,
				Limit:           10,
				ComponentTarget: &componentTargetFilter,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID:       1,
					ComponentTarget: &componentTargetFilter,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  10,
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "with all filters",
			input: &v1.ListSymbolsRequest{
				ProjectId:       1,
				Offset:          5,
				Limit:           25,
				Label:           &labelFilter,
				ComponentTarget: &componentTargetFilter,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID:       1,
					Label:           &labelFilter,
					ComponentTarget: &componentTargetFilter,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 5,
					Limit:  25,
				},
				Sort: domain.DefaultSortOption(),
			},
		},
		{
			name: "includes default sort option",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     10,
			},
			expected: domain.ListSymbolsOptions{
				Filter: domain.SymbolFilter{
					ProjectID: 1,
				},
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  10,
				},
				Sort: domain.SortOption{
					Field:     domain.SortByID,
					Direction: domain.SortAsc,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewListSymbolsOptions(tt.input)

			assert.Equal(t, tt.expected.Filter.ProjectID, result.Filter.ProjectID)
			assert.Equal(t, tt.expected.Filter.Label, result.Filter.Label)
			assert.Equal(t, tt.expected.Filter.ComponentTarget, result.Filter.ComponentTarget)
			assert.Equal(t, tt.expected.Pagination.Offset, result.Pagination.Offset)
			assert.Equal(t, tt.expected.Pagination.Limit, result.Pagination.Limit)
			assert.Equal(t, tt.expected.Sort.Field, result.Sort.Field)
			assert.Equal(t, tt.expected.Sort.Direction, result.Sort.Direction)
		})
	}
}
