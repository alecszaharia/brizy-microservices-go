// Package event provides mappings between domain events and protobuf events.
package event

import (
	"time"

	eventsv1 "contracts/gen/events/symbols/v1"
	"symbols/internal/biz/domain"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToSymbolCreatedEvent converts a domain.Symbol to a SymbolCreated event.
func ToSymbolCreatedEvent(s *domain.Symbol, createdAt time.Time) *eventsv1.SymbolCreated {
	if s == nil {
		return nil
	}
	return &eventsv1.SymbolCreated{
		Id:              s.ID,
		ProjectId:       s.Project,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		CreatedAt:       timestamppb.New(createdAt),
	}
}

// ToSymbolUpdatedEvent converts a domain.Symbol to a SymbolUpdated event.
func ToSymbolUpdatedEvent(s *domain.Symbol, updatedAt time.Time) *eventsv1.SymbolUpdated {
	if s == nil {
		return nil
	}
	return &eventsv1.SymbolUpdated{
		Id:              s.ID,
		ProjectId:       s.Project,
		Uid:             s.UID,
		Label:           s.Label,
		ClassName:       s.ClassName,
		ComponentTarget: s.ComponentTarget,
		Version:         s.Version,
		UpdatedAt:       timestamppb.New(updatedAt),
	}
}

// ToSymbolDeletedEvent creates a SymbolDeleted event from a domain.Symbol.
func ToSymbolDeletedEvent(s *domain.Symbol, deletedAt time.Time) *eventsv1.SymbolDeleted {
	if s == nil {
		return nil
	}
	return &eventsv1.SymbolDeleted{
		Id:        s.ID,
		ProjectId: s.Project,
		Uid:       s.UID,
		DeletedAt: timestamppb.New(deletedAt),
	}
}
