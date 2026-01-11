package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockPublisher implements events.Publisher for testing
type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	args := m.Called(ctx, topic, payload)
	return args.Error(0)
}

func (m *mockPublisher) Unwrap() message.Publisher {
	return nil
}

// mockSubscriber implements events.Subscriber for testing
type mockSubscriber struct {
	mock.Mock
}

func (m *mockSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	args := m.Called(ctx, topic)
	if ch := args.Get(0); ch != nil {
		return ch.(<-chan *message.Message), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockSubscriber) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSubscriber) Unwrap() message.Subscriber {
	return nil
}

func TestNewPublisherWithMetrics(t *testing.T) {
	reg := NewRegistry("test_service")
	basePub := &mockPublisher{}

	pub := NewPublisherWithMetrics(basePub, reg)

	require.NotNil(t, pub)
	assert.IsType(t, &publisherWithMetrics{}, pub)
}

func TestNewPublisherWithMetrics_NilRegistry(t *testing.T) {
	basePub := &mockPublisher{}

	pub := NewPublisherWithMetrics(basePub, nil)

	assert.Equal(t, basePub, pub, "Should return original publisher when registry is nil")
}

func TestNewPublisherWithMetrics_NilPublisher(t *testing.T) {
	reg := NewRegistry("test_service")

	pub := NewPublisherWithMetrics(nil, reg)

	assert.Nil(t, pub, "Should return nil when publisher is nil")
}

func TestPublisherWithMetrics_Publish_Success(t *testing.T) {
	reg := NewRegistry("test_service")
	basePub := &mockPublisher{}

	basePub.On("Publish", mock.Anything, "test.topic", mock.Anything).Return(nil)

	pub := NewPublisherWithMetrics(basePub, reg)

	err := pub.Publish(context.Background(), "test.topic", []byte("payload"))
	require.NoError(t, err)

	basePub.AssertExpectations(t)

	// Check metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	var foundPublished, foundDuration bool
	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_published_total" {
			foundPublished = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Equal(t, float64(1), m.GetMetric()[0].GetCounter().GetValue())
			// Check topic label
			labels := m.GetMetric()[0].GetLabel()
			assert.Equal(t, "test.topic", labels[0].GetValue())
		}
		if m.GetName() == "test_service_watermill_publish_duration_seconds" {
			foundDuration = true
		}
	}

	assert.True(t, foundPublished, "Should have published counter")
	assert.True(t, foundDuration, "Should have publish duration histogram")
}

func TestPublisherWithMetrics_Publish_Error(t *testing.T) {
	reg := NewRegistry("test_service")
	basePub := &mockPublisher{}

	publishErr := errors.New("publish failed")
	basePub.On("Publish", mock.Anything, "test.topic", mock.Anything).Return(publishErr)

	pub := NewPublisherWithMetrics(basePub, reg)

	err := pub.Publish(context.Background(), "test.topic", []byte("payload"))
	require.Error(t, err)
	assert.Equal(t, publishErr, err)

	basePub.AssertExpectations(t)

	// Check error metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	var foundError bool
	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_publish_errors_total" {
			foundError = true
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Equal(t, float64(1), m.GetMetric()[0].GetCounter().GetValue())
		}
		// Should not increment published counter on error
		if m.GetName() == "test_service_watermill_published_total" {
			assert.Equal(t, 0, len(m.GetMetric()))
		}
	}

	assert.True(t, foundError, "Should have error counter")
}

func TestPublisherWithMetrics_MultipleTopics(t *testing.T) {
	reg := NewRegistry("test_service")
	basePub := &mockPublisher{}

	basePub.On("Publish", mock.Anything, "topic.one", mock.Anything).Return(nil)
	basePub.On("Publish", mock.Anything, "topic.two", mock.Anything).Return(nil)

	pub := NewPublisherWithMetrics(basePub, reg)

	err := pub.Publish(context.Background(), "topic.one", []byte("payload1"))
	require.NoError(t, err)

	err = pub.Publish(context.Background(), "topic.two", []byte("payload2"))
	require.NoError(t, err)

	// Check metrics for both topics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_published_total" {
			assert.Equal(t, 2, len(m.GetMetric()))
		}
	}
}

func TestNewSubscriberWithMetrics(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	sub := NewSubscriberWithMetrics(baseSub, reg)

	require.NotNil(t, sub)
	assert.IsType(t, &subscriberWithMetrics{}, sub)
}

func TestNewSubscriberWithMetrics_NilRegistry(t *testing.T) {
	baseSub := &mockSubscriber{}

	sub := NewSubscriberWithMetrics(baseSub, nil)

	assert.Equal(t, baseSub, sub, "Should return original subscriber when registry is nil")
}

func TestSubscriberWithMetrics_Subscribe_Success(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	// Create a channel with messages
	msgCh := make(chan *message.Message, 1)
	msg := message.NewMessage("1", []byte("payload"))
	msgCh <- msg
	close(msgCh)

	baseSub.On("Subscribe", mock.Anything, "test.topic").Return((<-chan *message.Message)(msgCh), nil)

	sub := NewSubscriberWithMetrics(baseSub, reg)

	ch, err := sub.Subscribe(context.Background(), "test.topic")
	require.NoError(t, err)
	require.NotNil(t, ch)

	// Read message and ack it
	receivedMsg := <-ch
	require.NotNil(t, receivedMsg)
	receivedMsg.Ack()

	// Wait for metrics to be recorded
	time.Sleep(10 * time.Millisecond)

	// Check metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	var foundConsumed, foundAcks bool
	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_consumed_total" {
			foundConsumed = true
			if len(m.GetMetric()) > 0 {
				assert.Equal(t, float64(1), m.GetMetric()[0].GetCounter().GetValue())
			}
		}
		if m.GetName() == "test_service_watermill_handler_acks_total" {
			foundAcks = true
		}
	}

	assert.True(t, foundConsumed, "Should have consumed counter")
	assert.True(t, foundAcks, "Should have acks counter")

	baseSub.AssertExpectations(t)
}

func TestSubscriberWithMetrics_Subscribe_Error(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	subscribeErr := errors.New("subscribe failed")
	baseSub.On("Subscribe", mock.Anything, "test.topic").Return(nil, subscribeErr)

	sub := NewSubscriberWithMetrics(baseSub, reg)

	ch, err := sub.Subscribe(context.Background(), "test.topic")
	require.Error(t, err)
	assert.Nil(t, ch)
	assert.Equal(t, subscribeErr, err)

	// Check error metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_consume_errors_total" {
			assert.Equal(t, 1, len(m.GetMetric()))
			assert.Equal(t, float64(1), m.GetMetric()[0].GetCounter().GetValue())
		}
	}

	baseSub.AssertExpectations(t)
}

func TestSubscriberWithMetrics_Nack(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	msgCh := make(chan *message.Message, 1)
	msg := message.NewMessage("1", []byte("payload"))
	msgCh <- msg
	close(msgCh)

	baseSub.On("Subscribe", mock.Anything, "test.topic").Return((<-chan *message.Message)(msgCh), nil)

	sub := NewSubscriberWithMetrics(baseSub, reg)

	ch, err := sub.Subscribe(context.Background(), "test.topic")
	require.NoError(t, err)

	// Read message and nack it
	receivedMsg := <-ch
	require.NotNil(t, receivedMsg)
	receivedMsg.Nack()

	// Wait for metrics to be recorded
	time.Sleep(10 * time.Millisecond)

	// Check nack metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_watermill_handler_nacks_total" {
			if len(m.GetMetric()) > 0 {
				assert.Greater(t, m.GetMetric()[0].GetCounter().GetValue(), float64(0))
			}
		}
	}
}

func TestSubscriberWithMetrics_Close(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	baseSub.On("Close").Return(nil)

	sub := NewSubscriberWithMetrics(baseSub, reg)

	err := sub.Close()
	require.NoError(t, err)

	baseSub.AssertExpectations(t)
}

func TestSubscriberWithMetrics_Close_Error(t *testing.T) {
	reg := NewRegistry("test_service")
	baseSub := &mockSubscriber{}

	closeErr := errors.New("close failed")
	baseSub.On("Close").Return(closeErr)

	sub := NewSubscriberWithMetrics(baseSub, reg)

	err := sub.Close()
	require.Error(t, err)
	assert.Equal(t, closeErr, err)

	baseSub.AssertExpectations(t)
}

func TestPublisherWithMetrics_Unwrap(t *testing.T) {
	basePub := &mockPublisher{}
	basePub.On("Unwrap").Return(nil)

	reg := NewRegistry("test_service")
	pub := NewPublisherWithMetrics(basePub, reg).(*publisherWithMetrics)

	unwrapped := pub.Unwrap()
	assert.Nil(t, unwrapped)
}

func TestSubscriberWithMetrics_Unwrap(t *testing.T) {
	baseSub := &mockSubscriber{}
	baseSub.On("Unwrap").Return(nil)

	reg := NewRegistry("test_service")
	sub := NewSubscriberWithMetrics(baseSub, reg).(*subscriberWithMetrics)

	unwrapped := sub.Unwrap()
	assert.Nil(t, unwrapped)
}
