// Package common provides shared utilities for data layer operations including transaction management.
package common //nolint:revive // "common" is an acceptable name for shared utilities

import (
	"context"
)

type Transaction interface {
	InTx(context.Context, func(ctx context.Context) error) error
}
