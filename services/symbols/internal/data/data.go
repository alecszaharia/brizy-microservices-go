package data

import (
	"context"
	"symbols/internal/conf/gen"
	"symbols/internal/data/common"
	"symbols/internal/data/model"
	"symbols/internal/data/mq"
	"symbols/internal/data/repo"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is Data providers.
var ProviderSet = wire.NewSet(NewDB, NewData, NewTransaction, NewAmqpPublisher, repo.NewSymbolRepo, mq.NewEventPublisher)

// Data .
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

func NewAmqpPublisher(cfg *conf.Data, logger log.Logger) *amqp.Publisher {
	amqpConfig := amqp.NewDurablePubSubConfig(
		cfg.Mq.Addr,
		amqp.GenerateQueueNameTopicNameWithSuffix(cfg.Mq.Queue.Name),
	)

	amqpConfig.Exchange = amqp.ExchangeConfig{
		GenerateName: func(topic string) string {
			return cfg.Mq.Exchange.Name
		},
		Type:        cfg.Mq.Exchange.Type,
		Durable:     cfg.Mq.Exchange.Durable.GetValue(),
		AutoDeleted: cfg.Mq.Exchange.AutoDelete.GetValue(),
	}

	publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(false, false))
	if err != nil {
		log.NewHelper(logger).Errorf("failed to create AMQP publisher: %v", err)
	}
	return publisher
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}
