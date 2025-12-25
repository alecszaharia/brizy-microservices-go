package data

import (
	"context"
	"symbols/internal/conf"
	"symbols/internal/data/common"
	"symbols/internal/data/model"
	"symbols/internal/data/repo"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewDB, NewData, NewTransaction, repo.NewSymbolRepo)

// data .
type data struct {
	db  *gorm.DB
	log *log.Helper
}

// NewData .
func NewData(db *gorm.DB, logger log.Logger) (*data, func(), error) {
	l := log.NewHelper(logger)
	cleanup := func() {
		sqlDB, _ := db.DB()
		log.NewHelper(logger).Info("closing the data resources")
		err := sqlDB.Close()
		if err != nil {
			log.NewHelper(logger).Errorf("failed to close the data resource")
			return
		}
	}

	return &data{db: db, log: l}, cleanup, nil
}

func NewDB(cfg *conf.Data, logger log.Logger) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.Database.Source), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", cfg.Database.Source)
	}

	if cfg.Database.GetRunMigrations() {
		if err := db.AutoMigrate(&model.Symbol{}, &model.SymbolData{}); err != nil {
			log.Fatal(err)
		}
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(int(cfg.Database.MaxOpenConns))
	sqlDB.SetMaxIdleConns(int(cfg.Database.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime.AsDuration())

	log.NewHelper(logger).Info("database connection established")

	return db
}

// NewTransaction provider
func NewTransaction(d *data) common.Transaction {
	return d
}

type contextTxKey struct{}

func (d *data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}
