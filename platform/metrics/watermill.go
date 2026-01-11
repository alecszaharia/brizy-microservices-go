package metrics

import (
	"context"
	"platform/events"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/prometheus/client_golang/prometheus"
)

// publisherWithMetrics wraps an events.Publisher with metrics collection.
type publisherWithMetrics struct {
	publisher         events.Publisher
	publishedTotal    *prometheus.CounterVec
	publishDuration   *prometheus.HistogramVec
	publishErrorTotal *prometheus.CounterVec
}

// NewPublisherWithMetrics wraps a publisher with metrics collection.
func NewPublisherWithMetrics(pub events.Publisher, registry *Registry) events.Publisher {
	if registry == nil || pub == nil {
		return pub
	}

	return &publisherWithMetrics{
		publisher: pub,
		publishedTotal: registry.NewCounterVec(
			"watermill_published_total",
			"Total number of messages published",
			[]string{"topic"},
		),
		publishDuration: registry.NewHistogramVec(
			"watermill_publish_duration_seconds",
			"Message publish duration in seconds",
			defaultHistogramBuckets,
			[]string{"topic"},
		),
		publishErrorTotal: registry.NewCounterVec(
			"watermill_publish_errors_total",
			"Total number of publish errors",
			[]string{"topic"},
		),
	}
}

func (p *publisherWithMetrics) Publish(ctx context.Context, topic string, payload []byte) error {
	startTime := time.Now()

	err := p.publisher.Publish(ctx, topic, payload)

	duration := time.Since(startTime).Seconds()
	p.publishDuration.WithLabelValues(topic).Observe(duration)

	if err != nil {
		p.publishErrorTotal.WithLabelValues(topic).Inc()
	} else {
		p.publishedTotal.WithLabelValues(topic).Inc()
	}

	return err
}

func (p *publisherWithMetrics) Unwrap() message.Publisher {
	return p.publisher.Unwrap()
}

// subscriberWithMetrics wraps an events.Subscriber with metrics collection.
type subscriberWithMetrics struct {
	subscriber        events.Subscriber
	consumedTotal     *prometheus.CounterVec
	consumeDuration   *prometheus.HistogramVec
	consumeErrorTotal *prometheus.CounterVec
	handlerAcksTotal  *prometheus.CounterVec
	handlerNacksTotal *prometheus.CounterVec
}

// NewSubscriberWithMetrics wraps a subscriber with metrics collection.
func NewSubscriberWithMetrics(sub events.Subscriber, registry *Registry) events.Subscriber {
	if registry == nil || sub == nil {
		return sub
	}

	return &subscriberWithMetrics{
		subscriber: sub,
		consumedTotal: registry.NewCounterVec(
			"watermill_consumed_total",
			"Total number of messages consumed",
			[]string{"topic"},
		),
		consumeDuration: registry.NewHistogramVec(
			"watermill_consume_duration_seconds",
			"Message consume duration in seconds",
			defaultHistogramBuckets,
			[]string{"topic"},
		),
		consumeErrorTotal: registry.NewCounterVec(
			"watermill_consume_errors_total",
			"Total number of consume errors",
			[]string{"topic"},
		),
		handlerAcksTotal: registry.NewCounterVec(
			"watermill_handler_acks_total",
			"Total number of acknowledged messages",
			[]string{"topic"},
		),
		handlerNacksTotal: registry.NewCounterVec(
			"watermill_handler_nacks_total",
			"Total number of nacked messages",
			[]string{"topic"},
		),
	}
}

func (s *subscriberWithMetrics) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		s.consumeErrorTotal.WithLabelValues(topic).Inc()
		return nil, err
	}

	// Wrap the message channel to collect metrics
	wrappedMessages := make(chan *message.Message)

	go func() {
		defer close(wrappedMessages)

		for msg := range messages {
			startTime := time.Now()

			// Track message consumption
			s.consumedTotal.WithLabelValues(topic).Inc()

			// Forward the message to the handler
			wrappedMessages <- msg

			// Wait for the message to be acked or nacked
			select {
			case <-msg.Acked():
				duration := time.Since(startTime).Seconds()
				s.consumeDuration.WithLabelValues(topic).Observe(duration)
				s.handlerAcksTotal.WithLabelValues(topic).Inc()
			case <-msg.Nacked():
				duration := time.Since(startTime).Seconds()
				s.consumeDuration.WithLabelValues(topic).Observe(duration)
				s.handlerNacksTotal.WithLabelValues(topic).Inc()
			case <-ctx.Done():
				// Context cancelled, stop processing
				return
			}
		}
	}()

	return wrappedMessages, nil
}

func (s *subscriberWithMetrics) Close() error {
	return s.subscriber.Close()
}

func (s *subscriberWithMetrics) Unwrap() message.Subscriber {
	return s.subscriber.Unwrap()
}
