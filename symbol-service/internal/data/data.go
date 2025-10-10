package data

import (
	"symbol-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewSymbolRepo)

// Data .
type Data struct {
	db *gorm.DB
}

func (d *Data) Migrate() error {
	return d.db.AutoMigrate(&SymbolEntity{})
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}

	return &Data{db: db}, cleanup, nil
}
