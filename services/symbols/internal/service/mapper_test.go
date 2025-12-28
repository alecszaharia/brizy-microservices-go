package service

import (
	v1 "contracts/gen/symbols/v1"
	"platform/pagination"
	"testing"

	"symbols/internal/biz"

	"github.com/stretchr/testify/assert"
)

func Test_toBizSymbol(t *testing.T) {
	data := []byte(`{"key": "value"}`)

	tests := []struct {
		name     string
		input    *v1.Symbol
		expected *biz.Symbol
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
			expected: &biz.Symbol{
				Id:              123,
				Project:         456,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "web",
				Version:         1,
				Data: &biz.SymbolData{
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
			expected: &biz.Symbol{
				Id:              1,
				Project:         2,
				Uid:             "550e8400-e29b-41d4-a716-446655440001",
				Label:           "Empty Data Symbol",
				ClassName:       "EmptyClass",
				ComponentTarget: "mobile",
				Version:         1,
				Data: &biz.SymbolData{
					Project: 2,
					Data:    &[]byte{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBizSymbol(tt.input)

			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Project, result.Project)
			assert.Equal(t, tt.expected.Uid, result.Uid)
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
		expected *biz.Symbol
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
			expected: &biz.Symbol{
				Project:         789,
				Uid:             "550e8400-e29b-41d4-a716-446655440002",
				Label:           "New Symbol",
				ClassName:       "NewClass",
				ComponentTarget: "desktop",
				Version:         1,
				Data: &biz.SymbolData{
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
			expected: &biz.Symbol{
				Project:         1,
				Uid:             "550e8400-e29b-41d4-a716-446655440003",
				Label:           "Min Symbol",
				ClassName:       "MinClass",
				ComponentTarget: "app",
				Version:         1,
				Data: &biz.SymbolData{
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
			assert.Equal(t, tt.expected.Uid, result.Uid)
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
		input    *biz.Symbol
		expected *v1.Symbol
	}{
		{
			name: "complete biz symbol to v1",
			input: &biz.Symbol{
				Id:              999,
				Project:         888,
				Uid:             "550e8400-e29b-41d4-a716-446655440004",
				Label:           "Converted Symbol",
				ClassName:       "ConvertedClass",
				ComponentTarget: "universal",
				Version:         1,
				Data: &biz.SymbolData{
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
			input: &biz.Symbol{
				Id:              100,
				Project:         200,
				Uid:             "550e8400-e29b-41d4-a716-446655440005",
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
			input: &biz.Symbol{
				Id:              101,
				Project:         201,
				Uid:             "550e8400-e29b-41d4-a716-446655440006",
				Label:           "Empty Bytes Symbol",
				ClassName:       "EmptyBytesClass",
				ComponentTarget: "all",
				Version:         1,
				Data: &biz.SymbolData{
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
	tests := []struct {
		name     string
		input    *v1.ListSymbolsRequest
		wantErr  bool
		expected *biz.ListSymbolsOptions
	}{
		{
			name: "with offset and limit",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    10,
				Limit:     20,
			},
			wantErr: false,
			expected: &biz.ListSymbolsOptions{
				ProjectID: 1,
				Pagination: pagination.OffsetPaginationParams{
					Offset: 10,
					Limit:  20,
				},
			},
		},
		{
			name: "with default limit (zero)",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     0,
			},
			wantErr: false,
			expected: &biz.ListSymbolsOptions{
				ProjectID: 1,
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  20, // Default value
				},
			},
		},
		{
			name: "first page",
			input: &v1.ListSymbolsRequest{
				ProjectId: 1,
				Offset:    0,
				Limit:     10,
			},
			wantErr: false,
			expected: &biz.ListSymbolsOptions{
				ProjectID: 1,
				Pagination: pagination.OffsetPaginationParams{
					Offset: 0,
					Limit:  10,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewListSymbolsOptions(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ProjectID, result.ProjectID)
				assert.Equal(t, tt.expected.Pagination.Offset, result.Pagination.Offset)
				assert.Equal(t, tt.expected.Pagination.Limit, result.Pagination.Limit)
			}
		})
	}
}
