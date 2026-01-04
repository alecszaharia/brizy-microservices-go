package events

import "context"

// Publisher publishes events to an external messaging system.
type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// EventHandler handles events from an external messaging system.
type EventHandler interface {
	Handle(ctx context.Context, payload []byte) error
}
