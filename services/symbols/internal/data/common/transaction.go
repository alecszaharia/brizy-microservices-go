// Package common provides shared utilities for data layer operations including transaction management.
package common //nolint:revive // "common" is an acceptable name for shared utilities

import (
	"context"

	"gorm.io/gorm"
)

type Transaction interface {
	InTx(context.Context, func(ctx context.Context, tx *gorm.DB) error) error
}
