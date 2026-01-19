package biz

import (
	"github.com/go-playground/validator/v10"
)

// ProviderSet is biz providers.

func NewSymbolValidator() *validator.Validate {
	v := validator.New()

	// add all custom validations here
	// https://github.com/go-playground/validator/blob/master/_examples/struct-level/main.go.
	// https://github.com/go-playground/validator/blob/master/_examples/simple/main.go
	return v
}
