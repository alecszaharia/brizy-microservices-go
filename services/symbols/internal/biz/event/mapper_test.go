// Package event provides tests for mappings between domain events and protobuf events.
package event

import (
	"testing"
	"time"

	eventsv1 "contracts/gen/events/symbols/v1"
	"symbols/internal/biz/domain"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToSymbolCreatedEvent(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		symbol    *domain.Symbol
		createdAt time.Time
		expected  *eventsv1.SymbolCreated
	}{
		{
			name: "complete symbol conversion",
			symbol: &domain.Symbol{
				ID:              123,
				Project:         456,
				UID:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "web",
				Version:         1,
			},
			createdAt: createdAt,
			expected: &eventsv1.SymbolCreated{
				Id:              123,
				ProjectId:       456,
				Uid:             "550e8400-e29b-41d4-a716-446655440000",
				Label:           "Test Symbol",
				ClassName:       "TestClass",
				ComponentTarget: "web",
				Version:         1,
				CreatedAt:       timestamppb.New(createdAt),
			},
		},
		{
			name: "symbol with zero values",
			symbol: &domain.Symbol{
				ID:              0,
				Project:         1,
				UID:             "550e8400-e29b-41d4-a716-446655440001",
				Label:           "Min Symbol",
				ClassName:       "MinClass",
				ComponentTarget: "mobile",
				Version:         0,
			},
			createdAt: createdAt,
			expected: &eventsv1.SymbolCreated{
				Id:              0,
				ProjectId:       1,
				Uid:             "550e8400-e29b-41d4-a716-446655440001",
				Label:           "Min Symbol",
				ClassName:       "MinClass",
				ComponentTarget: "mobile",
				Version:         0,
				CreatedAt:       timestamppb.New(createdAt),
			},
		},
		{
			name:      "nil symbol returns nil",
			symbol:    nil,
			createdAt: createdAt,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSymbolCreatedEvent(tt.symbol, tt.createdAt)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.ProjectId, result.ProjectId)
			assert.Equal(t, tt.expected.Uid, result.Uid)
			assert.Equal(t, tt.expected.Label, result.Label)
			assert.Equal(t, tt.expected.ClassName, result.ClassName)
			assert.Equal(t, tt.expected.ComponentTarget, result.ComponentTarget)
			assert.Equal(t, tt.expected.Version, result.Version)
			assert.True(t, tt.expected.CreatedAt.AsTime().Equal(result.CreatedAt.AsTime()))
		})
	}
}

func TestToSymbolUpdatedEvent(t *testing.T) {
	updatedAt := time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC)

	tests := []struct {
		name      string
		symbol    *domain.Symbol
		updatedAt time.Time
		expected  *eventsv1.SymbolUpdated
	}{
		{
			name: "complete symbol conversion",
			symbol: &domain.Symbol{
				ID:              789,
				Project:         101,
				UID:             "550e8400-e29b-41d4-a716-446655440002",
				Label:           "Updated Symbol",
				ClassName:       "UpdatedClass",
				ComponentTarget: "desktop",
				Version:         2,
			},
			updatedAt: updatedAt,
			expected: &eventsv1.SymbolUpdated{
				Id:              789,
				ProjectId:       101,
				Uid:             "550e8400-e29b-41d4-a716-446655440002",
				Label:           "Updated Symbol",
				ClassName:       "UpdatedClass",
				ComponentTarget: "desktop",
				Version:         2,
				UpdatedAt:       timestamppb.New(updatedAt),
			},
		},
		{
			name: "symbol with updated version",
			symbol: &domain.Symbol{
				ID:              1,
				Project:         2,
				UID:             "550e8400-e29b-41d4-a716-446655440003",
				Label:           "Version Bump",
				ClassName:       "VersionClass",
				ComponentTarget: "app",
				Version:         5,
			},
			updatedAt: updatedAt,
			expected: &eventsv1.SymbolUpdated{
				Id:              1,
				ProjectId:       2,
				Uid:             "550e8400-e29b-41d4-a716-446655440003",
				Label:           "Version Bump",
				ClassName:       "VersionClass",
				ComponentTarget: "app",
				Version:         5,
				UpdatedAt:       timestamppb.New(updatedAt),
			},
		},
		{
			name:      "nil symbol returns nil",
			symbol:    nil,
			updatedAt: updatedAt,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSymbolUpdatedEvent(tt.symbol, tt.updatedAt)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.ProjectId, result.ProjectId)
			assert.Equal(t, tt.expected.Uid, result.Uid)
			assert.Equal(t, tt.expected.Label, result.Label)
			assert.Equal(t, tt.expected.ClassName, result.ClassName)
			assert.Equal(t, tt.expected.ComponentTarget, result.ComponentTarget)
			assert.Equal(t, tt.expected.Version, result.Version)
			assert.True(t, tt.expected.UpdatedAt.AsTime().Equal(result.UpdatedAt.AsTime()))
		})
	}
}

func TestToSymbolDeletedEvent(t *testing.T) {
	deletedAt := time.Date(2024, 3, 25, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		symbol    *domain.Symbol
		deletedAt time.Time
		expected  *eventsv1.SymbolDeleted
	}{
		{
			name: "complete symbol conversion",
			symbol: &domain.Symbol{
				ID:              555,
				Project:         666,
				UID:             "550e8400-e29b-41d4-a716-446655440004",
				Label:           "To Delete",
				ClassName:       "DeleteClass",
				ComponentTarget: "all",
				Version:         3,
			},
			deletedAt: deletedAt,
			expected: &eventsv1.SymbolDeleted{
				Id:        555,
				ProjectId: 666,
				Uid:       "550e8400-e29b-41d4-a716-446655440004",
				DeletedAt: timestamppb.New(deletedAt),
			},
		},
		{
			name: "symbol with minimal fields",
			symbol: &domain.Symbol{
				ID:      1,
				Project: 2,
				UID:     "550e8400-e29b-41d4-a716-446655440005",
			},
			deletedAt: deletedAt,
			expected: &eventsv1.SymbolDeleted{
				Id:        1,
				ProjectId: 2,
				Uid:       "550e8400-e29b-41d4-a716-446655440005",
				DeletedAt: timestamppb.New(deletedAt),
			},
		},
		{
			name:      "nil symbol returns nil",
			symbol:    nil,
			deletedAt: deletedAt,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSymbolDeletedEvent(tt.symbol, tt.deletedAt)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.ProjectId, result.ProjectId)
			assert.Equal(t, tt.expected.Uid, result.Uid)
			assert.True(t, tt.expected.DeletedAt.AsTime().Equal(result.DeletedAt.AsTime()))
		})
	}
}
