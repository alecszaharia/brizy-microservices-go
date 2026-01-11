package mq

import (
	"context"
	"platform/events"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
)

func NewEventSubscriber(sub message.Subscriber, logger log.Logger) events.Subscriber {
	return &eventSubscriber{
		sub:    sub,
		logger: log.NewHelper(logger),
	}
}

type eventSubscriber struct {
	sub    message.Subscriber
	logger *log.Helper
}

func (es *eventSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	es.logger.WithContext(ctx).Infof("Subscribing to topic: %s", topic)

	messages, err := es.sub.Subscribe(ctx, topic)
	if err != nil {
		es.logger.WithContext(ctx).Errorf("Failed to subscribe to topic %s: %v", topic, err)
		return nil, err
	}

	es.logger.WithContext(ctx).Infof("Successfully subscribed to topic: %s", topic)
	return messages, nil
}

func (es *eventSubscriber) Close() error {
	es.logger.Info("Closing event subscriber")

	if err := es.sub.Close(); err != nil {
		es.logger.Errorf("Failed to close subscriber: %v", err)
		return err
	}

	es.logger.Info("Event subscriber closed successfully")
	return nil
}

// Unwrap returns the underlying AMQP subscriber for direct Watermill router usage
func (es *eventSubscriber) Unwrap() message.Subscriber {
	return es.sub
}
