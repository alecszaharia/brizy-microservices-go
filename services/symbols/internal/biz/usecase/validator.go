// Package usecase Package symbol provides validation for symbols.
package usecase

import (
	"github.com/go-playground/validator/v10"
)

// NewValidator creates a new validator instance for Symbol validation.
func NewValidator() *validator.Validate {
	v := validator.New()

	// add all custom validations here
	// https://github.com/go-playground/validator/blob/master/_examples/struct-level/main.go.
	// https://github.com/go-playground/validator/blob/master/_examples/simple/main.go
	return v
}
