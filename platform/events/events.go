// Package events provides interfaces for publishing and subscribing to events.
package events

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Publisher publishes events to an external messaging system.
type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
	Unwrap() message.Publisher
}

// Subscriber subscribes to events from an external messaging system.
// Returns a channel of messages for consumption.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	Close() error
	Unwrap() message.Subscriber
}

// EventHandler handles events from an external messaging system.
type EventHandler interface {
	Handle(ctx context.Context, payload []byte) error
}
