// Package data implements the data access layer with repositories and database connections.
package data

import (
	"context"
	"platform/events"
	platform_logger "platform/logger"
	"platform/metrics"
	"symbols/internal/conf/gen"
	"symbols/internal/data/common"
	"symbols/internal/data/model"
	"symbols/internal/data/mq"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Data provides database access and transaction management.
type Data struct {
	db *gorm.DB
}

// NewData initializes a new Data instance, a cleanup	 function, and returns an error if initialization fails.
func NewData(db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			l.Errorf("failed to get database handle during cleanup: %v", err)
			return // Exit early, don't proceed to Close
		}
		l.Info("closing the data resources")
		if err := sqlDB.Close(); err != nil {
			l.Errorf("failed to close the data resource: %v", err)
		}

	}

	return &Data{db: db}, cleanup, nil
}

// NewDB initializes and returns a new gorm.DB connection configured with the given settings and logger.
func NewDB(cfg *conf.Data, logger log.Logger) *gorm.DB {
	l := log.NewHelper(logger)
	db, err := gorm.Open(mysql.Open(cfg.Database.Source), &gorm.Config{})
	if err != nil {
		l.Fatalf("failed opening connection to mysql: %v", err)
	}

	if cfg.Database.RunMigrations.Value {
		if err := db.AutoMigrate(&model.Symbol{}, &model.SymbolData{}); err != nil {
			l.Fatalf("Failed to migrate: %v", err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		l.Fatalf("failed to get database handle: %v", err)
	}

	sqlDB.SetMaxOpenConns(int(cfg.Database.MaxOpenConns))
	sqlDB.SetMaxIdleConns(int(cfg.Database.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime.AsDuration())

	l.Info("database connection established")

	return db
}

// NewTransaction creates and returns a new Transaction instance using the provided Data dependency.
func NewTransaction(d *Data) common.Transaction {
	return d
}

func NewAMQPPublisher(cfg *conf.Data, logger log.Logger, wmLogger *platform_logger.WatermillLogger) message.Publisher {
	amqpConfig := amqp.NewDurablePubSubConfig(
		cfg.Mq.Addr,
		amqp.GenerateQueueNameConstant(cfg.Mq.Queue.Name),
	)

	amqpConfig.Exchange = amqp.ExchangeConfig{
		GenerateName: func(topic string) string {
			return cfg.Mq.Exchange.Name
		},
		Type:        cfg.Mq.Exchange.Type,
		Durable:     cfg.Mq.Exchange.Durable.Value,
		AutoDeleted: cfg.Mq.Exchange.AutoDelete.Value,
	}
	amqpConfig.Publish = amqp.PublishConfig{
		GenerateRoutingKey: func(topic string) string {
			return topic
		},
	}

	publisher, err := amqp.NewPublisher(amqpConfig, wmLogger)

	if err != nil {
		log.NewHelper(logger).Errorf("failed to create AMQP publisher: %v", err)
	}
	return publisher
}

func NewAMQPSubscriber(cfg *conf.Data, logger log.Logger, wmLogger *platform_logger.WatermillLogger) message.Subscriber {
	amqpConfig := amqp.NewDurablePubSubConfig(
		cfg.Mq.Addr,
		amqp.GenerateQueueNameConstant(cfg.Mq.Queue.Name),
	)

	amqpConfig.Queue = amqp.QueueConfig{
		GenerateName: func(topic string) string {
			return cfg.Mq.Queue.Name
		},
		Durable:    cfg.Mq.Queue.Durable.Value,
		AutoDelete: cfg.Mq.Queue.AutoDelete.Value,
		Exclusive:  cfg.Mq.Queue.Exclusive.Value,
		NoWait:     false,
	}

	amqpConfig.QueueBind = amqp.QueueBindConfig{
		GenerateRoutingKey: func(topic string) string {
			return cfg.Mq.Queue.BindingKey
		},
	}

	subscriber, err := amqp.NewSubscriber(amqpConfig, wmLogger)

	if err != nil {
		log.NewHelper(logger).Errorf("failed to create AMQP subscriber: %v", err)
	}
	return subscriber
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}

// NewEventPublisherWithMetrics wraps the base event publisher with metrics if enabled.
func NewEventPublisherWithMetrics(pub message.Publisher, mc *conf.Metrics, reg *metrics.Registry, logger log.Logger) events.Publisher {
	basePub := mq.NewEventPublisher(pub, logger)
	if mc != nil && mc.Enabled.Value && reg != nil {
		return metrics.NewPublisherWithMetrics(basePub, reg)
	}
	return basePub
}

// NewEventSubscriberWithMetrics wraps the base event subscriber with metrics if enabled.
func NewEventSubscriberWithMetrics(sub message.Subscriber, mc *conf.Metrics, reg *metrics.Registry, logger log.Logger) events.Subscriber {
	baseSub := mq.NewEventSubscriber(sub, logger)
	if mc != nil && mc.Enabled.Value && reg != nil {

		return metrics.NewSubscriberWithMetrics(baseSub, reg)
	}
	return baseSub
}
