// Package biz provides business logic providers.
package biz

import (
	"symbols/internal/biz/symbol"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(symbol.NewValidator, symbol.NewUseCase)
