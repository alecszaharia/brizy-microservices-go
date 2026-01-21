// Package mq provides message queue publisher and subscriber implementations.
package mq

import (
	"context"
	"fmt"
	middleware2 "platform/middleware"
	"symbols/internal/biz/domain"
	"symbols/internal/biz/event"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
)

const RoutingKey = "routing_key"

func NewEventPublisher(pub message.Publisher, logger log.Logger) domain.SymbolEventPublisher {
	return &eventPublisher{
		pub:    pub,
		logger: log.NewHelper(logger),
	}
}

type eventPublisher struct {
	pub    message.Publisher
	logger *log.Helper
}

// Unwrap returns the underlying AMQP publisher for direct Watermill router usage
func (ep *eventPublisher) Unwrap() message.Publisher {
	return ep.pub
}

func (ep *eventPublisher) Publish(ctx context.Context, topic string, payload []byte) error {

	msg := message.NewMessage(watermill.NewUUID(), payload)

	// Propagate context to subscriber
	msg.SetContext(ctx)

	// Set correlation ID if not already set
	correlationID := middleware.MessageCorrelationID(msg)
	if correlationID == "" {
		correlationID = watermill.NewUUID()
		middleware.SetCorrelationID(correlationID, msg)
	}

	// Set the routing key if not already set
	SetMessageRoutingKey(topic, msg)

	if requestID := extractRequestID(ctx); requestID != "" {
		msg.Metadata.Set(middleware2.RequestIDKey, requestID)
	}

	ep.logger.WithContext(ctx).Infof("Publishing message %s to topic %s, correlation_id: %s", msg.UUID, topic, correlationID)

	if err := ep.pub.Publish(topic, msg); err != nil {
		ep.logger.WithContext(ctx).Errorf("Failed to publish message %s to topic %s: %v", msg.UUID, topic, err)
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}

	return nil
}

func (ep *eventPublisher) PublishSymbolCreated(ctx context.Context, symbol *domain.Symbol) error {
	// Convert to event proto using mapper
	evt := event.ToSymbolCreatedEvent(symbol, time.Now())
	if evt == nil {
		return fmt.Errorf("failed to convert symbol to created event: symbol is nil")
	}

	// Marshal proto to bytes
	payload, err := proto.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal symbol created event: %w", err)
	}

	// Publish using base publisher (handles context, correlation ID, logging)
	return ep.Publish(ctx, event.SymbolCreatedTopic, payload)
}

func (ep *eventPublisher) PublishSymbolUpdated(ctx context.Context, symbol *domain.Symbol) error {
	// Convert to event proto using mapper
	evt := event.ToSymbolUpdatedEvent(symbol, time.Now())
	if evt == nil {
		return fmt.Errorf("failed to convert symbol to updated event: symbol is nil")
	}

	// Marshal proto to bytes
	payload, err := proto.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal symbol updated event: %w", err)
	}

	// Publish using base publisher (handles context, correlation ID, logging)
	return ep.Publish(ctx, event.SymbolUpdatedTopic, payload)
}

func (ep *eventPublisher) PublishSymbolDeleted(ctx context.Context, symbol *domain.Symbol) error {
	// Convert to event proto using mapper
	evt := event.ToSymbolDeletedEvent(symbol, time.Now())
	if evt == nil {
		return fmt.Errorf("failed to convert symbol to deleted event: symbol is nil")
	}

	// Marshal proto to bytes
	payload, err := proto.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal symbol deleted event: %w", err)
	}

	// Publish using base publisher (handles context, correlation ID, logging)
	return ep.Publish(ctx, event.SymbolDeletedTopic, payload)
}

func SetMessageRoutingKey(key string, msg *message.Message) {
	if MessageRoutingKey(msg) != "" {
		return
	}
	msg.Metadata.Set(RoutingKey, key)
}

func MessageRoutingKey(msg *message.Message) string {
	return msg.Metadata.Get(RoutingKey)
}

// extractRequestID safely extracts request ID from context
func extractRequestID(ctx context.Context) string {
	val := middleware2.RequestID()(ctx)
	if requestID, ok := val.(string); ok {
		return requestID
	}
	return ""
}
