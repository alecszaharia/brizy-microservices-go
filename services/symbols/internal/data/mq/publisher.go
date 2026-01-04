package mq

import (
	"context"
	"log"
	"symbols/internal/biz/events"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func NewEventPublisher(pub *amqp.Publisher) events.Publisher {
	return &eventPublisher{pub: pub}
}

type eventPublisher struct {
	pub *amqp.Publisher
}

func (ep *eventPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), payload)
	correlationId := watermill.NewUUID()
	middleware.SetCorrelationID(correlationId, msg)

	log.Printf("sending message %s, correlation id: %s\n", msg.UUID, correlationId)

	if err := ep.pub.Publish(topic, msg); err != nil {
		panic(err)
	}

	return nil
}
