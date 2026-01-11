# Event-Driven Architecture: Pub/Sub Implementation Guide

This guide explains the event-driven architecture implemented in the Brizy microservices platform, covering the publisher/subscriber pattern, context propagation, and how to extend the system with new event types and handlers.

## Architecture Overview

The pub/sub implementation is built on three layers:

1. **Platform Layer** (`platform/events`) - Abstract interfaces for any message broker
2. **Data Layer** (`services/*/internal/data/mq`) - Concrete AMQP/RabbitMQ implementations
3. **Application Layer** (`services/*/internal/handlers`) - Business logic event handlers

This layered approach allows swapping message brokers (AMQP, Redis, Kafka, SQS) without changing business logic.

## Core Concepts

### 1. Platform-Level Abstractions

**Location**: `platform/events/events.go`

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, payload []byte) error
    Unwrap() message.Publisher
}

type Subscriber interface {
    Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
    Close() error
    Unwrap() message.Subscriber
}
```

**Key Principles**:
- **Broker-agnostic** - Interfaces work with any message broker implementation
- **Context-aware** - All operations accept `context.Context` for tracing and cancellation
- **Unwrap pattern** - Access underlying broker-specific features when needed

### 2. Publisher Wrapper

**Location**: `services/symbols/internal/data/mq/publisher.go`

The publisher wrapper enriches outgoing messages with metadata:

**Automatic Enrichment**:
1. **Context Propagation** - Attaches context to message for subscriber chain
2. **Correlation ID** - Generates or preserves existing correlation ID
3. **Request ID** - Extracts from HTTP request context via middleware
4. **Routing Key** - Sets AMQP routing key from topic name
5. **Structured Logging** - Logs with full context (request_id, correlation_id)

**Example**:
```go
// Business layer publishes event
err := symbolUC.pub.Publish(ctx, "symbols.created", payload)

// Publisher wrapper automatically:
// - Adds msg.SetContext(ctx)
// - Adds correlation_id to metadata
// - Adds request_id from HTTP context
// - Logs with structured logger
```

### 3. Subscriber Wrapper

**Location**: `services/symbols/internal/data/mq/subscriber.go`

The subscriber wrapper provides lifecycle management and logging for message consumption.

### 4. Event Handlers

**Location**: `services/*/internal/handlers/*.go`

Handlers implement business logic for processing events:

```go
type LifecycleEventHandler struct {
    logger   *log.Helper
    symbolUC biz.SymbolUseCase  // Business layer dependency
}

func (h *LifecycleEventHandler) Handle(msg *message.Message) error {
    ctx := msg.Context()  // Extract context chain

    // Extract metadata
    correlationID := msg.Metadata.Get("correlation_id")
    requestID := msg.Metadata.Get("request_id")

    // Log with context (includes request_id automatically)
    h.logger.WithContext(ctx).Infof("Processing event...")

    // Delegate to business layer
    return h.symbolUC.ProcessEvent(ctx, payload)
}
```

**Handler Best Practices**:
- Always extract context: `ctx := msg.Context()`
- Always use `logger.WithContext(ctx)` for tracing
- Delegate business logic to use cases (biz layer)
- Return errors for retry (Watermill handles retry logic)

### 5. Worker Architecture

**Location**: `services/symbols/cmd/symbols-worker/`

The worker is a separate binary that runs the Watermill router:

**Structure**:
```
cmd/symbols-worker/
├── main.go          # Entry point, config loading
├── wire.go          # Dependency injection
└── wire_gen.go      # Generated Wire code

internal/worker/
├── provider.go      # Wire ProviderSet
└── worker.go        # Router setup, lifecycle management
```

**Worker Lifecycle**:
```go
// Kratos hooks integrate with Watermill router
kratos.BeforeStart(worker.Start())  // Starts router in goroutine
kratos.AfterStop(worker.Stop())     // Graceful shutdown with timeout
```

## Context Propagation Flow

Understanding context flow is critical for distributed tracing:

```
HTTP Request
  ↓
RequestIDMiddleware (platform/middleware/request_id.go)
  ↓ Injects request_id into context
Service Handler (internal/service)
  ↓ ctx contains request_id
Use Case (internal/biz)
  ↓ pub.Publish(ctx, topic, payload)
Publisher Wrapper (internal/data/mq)
  ↓ msg.SetContext(ctx) - Attaches context to message
  ↓ Extracts request_id from context
  ↓ Adds to message metadata
AMQP Broker
  ↓
Subscriber receives message
  ↓
Handler extracts context: ctx := msg.Context()
  ↓ ctx contains original request_id
Use Case processes event
  ↓ logger.WithContext(ctx) includes request_id in logs
```

**Result**: End-to-end tracing from HTTP request → event publishing → event consumption.

## Dependency Injection with Wire

The worker uses Wire to compose all dependencies:

```go
// cmd/symbols-worker/wire.go
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(
        worker.ProviderSet,     // Router, Worker
        handlers.ProviderSet,   // Event handlers
        data.ProviderSet,       // AMQP, DB, repos
        biz.ProviderSet,        // Use cases
        newApp,
    ))
}
```

**Layer Dependencies**:
```
worker.ProviderSet
  ├─ NewRouter(cfg, handlers, pub, sub, logger)
  └─ NewWorker(router, logger)

handlers.ProviderSet
  └─ NewLifecycleEventHandler(symbolUC, logger)
         └─ requires biz.SymbolUseCase

data.ProviderSet
  ├─ NewAmqpPublisher(cfg, logger)
  ├─ NewAmqpSubscriber(cfg, logger)
  └─ NewEventPublisher(amqpPub, logger)

biz.ProviderSet
  └─ NewSymbolUseCase(repo, validator, tx, pub, logger)
```

## Configuration

**Location**: `services/symbols/configs/config.yaml`

```yaml
data:
  mq:
    addr: ${MQ_ADDR:amqp://guest:guest@rabbitmq:5672/}
    exchange:
      name: ${MQ_EXCHANGE_NAME:lifecycle_events}
      type: ${MQ_EXCHANGE_TYPE:topic}  # Use 'topic' for routing keys
      durable: true
    queue:
      name: ${MQ_QUEUE_NAME:symbols_lifecycle_queue}
      durable: true
      binding_key: ${MQ_QUEUE_BINDING_KEY:symbols.#}  # Wildcard routing
      prefetch_count: ${MQ_QUEUE_PREFETCH_COUNT:10}
```

**Environment Variables**:
- Prefix with `KRATOS_` for Kratos config overrides
- Example: `KRATOS_DATA_MQ_ADDR=amqp://user:pass@localhost:5672/`

## Extending the System

### Adding a New Event Type

**Step 1**: Define event structure in business layer

```go
// services/symbols/internal/biz/events/types.go
package events

type SymbolCreatedEvent struct {
    SymbolID  uint64 `json:"symbol_id"`
    ProjectID uint64 `json:"project_id"`
    UserID    uint64 `json:"user_id,omitempty"`
    Timestamp int64  `json:"timestamp"`
}
```

**Step 2**: Publish from use case

```go
// services/symbols/internal/biz/symbols.go
func (uc *symbolUseCase) CreateSymbol(ctx context.Context, s *Symbol) (*Symbol, error) {
    // ... create symbol in database ...

    // Publish event
    event := events.SymbolCreatedEvent{
        SymbolID:  symbol.Id,
        ProjectID: symbol.ProjectId,
        Timestamp: time.Now().Unix(),
    }

    payload, _ := json.Marshal(event)
    if err := uc.pub.Publish(ctx, "symbols.created", payload); err != nil {
        uc.log.WithContext(ctx).Errorf("Failed to publish event: %v", err)
        // Decide: fail operation or log error and continue
    }

    return symbol, nil
}
```

**Step 3**: Create handler

```go
// services/symbols/internal/handlers/symbol_created.go
package handlers

type SymbolCreatedHandler struct {
    logger *log.Helper
    // Add dependencies (use cases, repos, etc.)
}

func NewSymbolCreatedHandler(logger log.Logger) *SymbolCreatedHandler {
    return &SymbolCreatedHandler{logger: log.NewHelper(logger)}
}

func (h *SymbolCreatedHandler) Handle(msg *message.Message) error {
    ctx := msg.Context()

    var event events.SymbolCreatedEvent
    if err := json.Unmarshal(msg.Payload, &event); err != nil {
        h.logger.WithContext(ctx).Errorf("Invalid event: %v", err)
        return err  // Will retry
    }

    h.logger.WithContext(ctx).Infof("Symbol %d created", event.SymbolID)

    // Process event (call use cases, update caches, etc.)
    return nil
}
```

**Step 4**: Register handler in router

```go
// services/symbols/internal/worker/worker.go
func NewRouter(
    cfg *conf.Data,
    symbolCreatedHandler *handlers.SymbolCreatedHandler,
    // ... other handlers
    sub *amqp.Subscriber,
    logger log.Logger,
) *message.Router {
    // ... router setup ...

    router.AddNoPublisherHandler(
        "symbol-created-handler",
        "symbols.created",
        sub,
        symbolCreatedHandler.Handle,
    )

    return router
}
```

**Step 5**: Update Wire providers

```go
// services/symbols/internal/handlers/provider.go
var ProviderSet = wire.NewSet(
    NewLifecycleEventHandler,
    NewSymbolCreatedHandler,  // Add new handler
)

// services/symbols/internal/worker/provider.go
var ProviderSet = wire.NewSet(NewRouter, NewWorker)
```

**Step 6**: Regenerate Wire code

```bash
cd services/symbols
make generate
```

### Adding a New Message Broker

To support Redis Pub/Sub, Kafka, or SQS:

**Step 1**: Create broker-specific subscriber/publisher

```go
// services/symbols/internal/data/mq/redis_publisher.go
package mq

import (
    "context"
    "platform/events"
    "github.com/redis/go-redis/v9"
)

func NewRedisPublisher(client *redis.Client, logger log.Logger) events.Publisher {
    return &redisPublisher{
        client: client,
        logger: log.NewHelper(logger),
    }
}

type redisPublisher struct {
    client *redis.Client
    logger *log.Helper
}

func (rp *redisPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
    // Implement Redis-specific publishing
    return rp.client.Publish(ctx, topic, payload).Err()
}

func (rp *redisPublisher) Unwrap() message.Publisher {
    return nil  // Or return Redis-specific type
}
```

**Step 2**: Add configuration

```yaml
data:
  redis:
    addr: ${REDIS_ADDR:localhost:6379}
    password: ${REDIS_PASSWORD:}
    db: ${REDIS_DB:0}
```

**Step 3**: Update Wire providers

```go
// Choose based on config or feature flag
func NewEventPublisher(cfg *conf.Data, logger log.Logger) events.Publisher {
    if cfg.UseBroker == "redis" {
        return mq.NewRedisPublisher(redisClient, logger)
    }
    return mq.NewEventPublisher(amqpPub, logger)
}
```

## Middleware and Enrichment

### Watermill Router Middleware

```go
// services/symbols/internal/worker/worker.go
router.AddMiddleware(
    middleware.CorrelationID,     // Propagates correlation_id
    middleware.Retry{             // Retries on error
        MaxRetries:      3,
        InitialInterval: 100 * time.Millisecond,
    }.Middleware,
    middleware.Recoverer,         // Catches panics
)
```

### Custom Middleware Example

```go
// Add custom middleware for metrics
func MetricsMiddleware(logger log.Logger) message.HandlerMiddleware {
    return func(h message.HandlerFunc) message.HandlerFunc {
        return func(msg *message.Message) ([]*message.Message, error) {
            start := time.Now()
            msgs, err := h(msg)
            duration := time.Since(start)

            // Log metrics
            log.NewHelper(logger).Infof(
                "Handler duration: %v, success: %v",
                duration,
                err == nil,
            )

            return msgs, err
        }
    }
}

// Register in router
router.AddMiddleware(MetricsMiddleware(logger))
```

## Troubleshooting

### Event Not Being Consumed

1. **Check queue binding**: Verify routing key pattern matches topic
   ```bash
   # RabbitMQ Management UI
   http://localhost:15672
   ```

2. **Check handler registration**: Ensure handler is registered in router
   ```go
   router.AddNoPublisherHandler("handler-name", "topic", sub, handler.Handle)
   ```

3. **Check logs**: Look for errors in worker logs
   ```bash
   docker-compose logs -f symbols-worker
   ```

### Missing Context Metadata

**Problem**: `request_id` not appearing in subscriber logs

**Solution**: Ensure publisher calls `msg.SetContext(ctx)`
```go
// Publisher must do this:
msg.SetContext(ctx)  // ← Critical!
```

### Exchange Type Mismatch

**Problem**: Messages not routed correctly

**Solution**: Use `topic` exchange for routing keys:
```yaml
exchange:
  type: topic  # Not 'fanout'
```

Routing patterns:
- `symbols.#` - Matches `symbols.created`, `symbols.updated`, etc.
- `symbols.created` - Matches only `symbols.created`
- `*.created` - Matches `symbols.created`, `projects.created`, etc.

## Best Practices

1. **Always use context**: Pass `context.Context` through the entire chain
2. **Always log with context**: Use `logger.WithContext(ctx)` for distributed tracing
3. **Keep payloads small**: Use IDs and let consumers fetch data if needed
4. **Version your events**: Add version field to event structures for evolution
5. **Handle failures gracefully**: Return errors for retries, log non-retriable errors
6. **Test handlers independently**: Mock use cases for unit testing handlers
7. **Monitor queue depth**: Alert when queues grow beyond expected size
8. **Document event schemas**: Maintain event catalog in documentation

## Testing

### Unit Test Publisher Wrapper

```go
func TestEventPublisher_Publish(t *testing.T) {
    mockPub := &mockPublisher{}
    logger := log.DefaultLogger
    pub := mq.NewEventPublisher(mockPub, logger)

    ctx := context.WithValue(context.Background(), "request_id", "test-123")
    err := pub.Publish(ctx, "test.topic", []byte("payload"))

    assert.NoError(t, err)
    assert.Equal(t, "test-123", mockPub.lastMsg.Metadata.Get("request_id"))
}
```

### Integration Test Event Flow

```go
func TestEventFlow_SymbolCreated(t *testing.T) {
    // Start test broker
    broker := setupTestBroker(t)
    defer broker.Close()

    // Publish event
    publisher := mq.NewEventPublisher(broker.Publisher(), logger)
    err := publisher.Publish(ctx, "symbols.created", payload)
    require.NoError(t, err)

    // Consume event
    subscriber := mq.NewEventSubscriber(broker.Subscriber(), logger)
    messages, _ := subscriber.Subscribe(ctx, "symbols.created")

    select {
    case msg := <-messages:
        assert.Equal(t, payload, msg.Payload)
    case <-time.After(5 * time.Second):
        t.Fatal("Event not received")
    }
}
```

## References

- [Watermill Documentation](https://watermill.io/docs/)
- [Go-Kratos Framework](https://go-kratos.dev/)
- [AMQP 0-9-1 Model](https://www.rabbitmq.com/tutorials/amqp-concepts.html)
- [Distributed Tracing Best Practices](https://opentelemetry.io/docs/concepts/observability-primer/)