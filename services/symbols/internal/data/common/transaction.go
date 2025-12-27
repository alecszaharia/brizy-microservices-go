package common

import (
	"context"

	"gorm.io/gorm"
)

type Transaction interface {
	InTx(context.Context, func(ctx context.Context, tx *gorm.DB) error) error
}
