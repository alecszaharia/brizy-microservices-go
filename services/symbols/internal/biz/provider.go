// Package biz provides business logic providers.
package biz

import (
	"symbols/internal/biz/usecase"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(usecase.NewValidator, usecase.NewUseCase)
