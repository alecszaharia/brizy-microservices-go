package domain

import "errors"

// Domain-level errors (returned by business logic layer)
var (
	// ErrSymbolNotFound is returned when a symbol is not found in the repository.
	// This should be returned from biz layer when the data layer returns ErrDataNotFound.
	ErrSymbolNotFound = errors.New("symbol not found")

	// ErrInvalidID is returned when a symbol ID is invalid (e.g., zero or negative).
	ErrInvalidID = errors.New("invalid symbol ID")

	// ErrValidationFailed is returned when symbol validation fails.
	ErrValidationFailed = errors.New("validation failed")

	// ErrDuplicateSymbol is returned when attempting to create a symbol that already exists.
	// This typically occurs when the unique constraint on (project_id, uid) is violated.
	ErrDuplicateSymbol = errors.New("symbol with this UID already exists in the project")

	// ErrDatabaseOperation is returned when a database operation fails.
	ErrDatabaseOperation = errors.New("database operation failed")
)

// Data layer errors (returned by repository implementations)
// These errors abstract away GORM-specific errors to maintain clean architecture
var (
	// ErrDataNotFound indicates the requested record does not exist in the database.
	// Data layer returns this instead of gorm.ErrRecordNotFound.
	ErrDataNotFound = errors.New("record not found")

	// ErrDataDuplicateEntry indicates a unique constraint violation.
	// Data layer returns this instead of MySQL duplicate key errors.
	ErrDataDuplicateEntry = errors.New("duplicate entry")

	// ErrDataTransactionFailed indicates a transaction operation failed.
	ErrDataTransactionFailed = errors.New("transaction failed")

	// ErrDataDatabase indicates a generic database operation error.
	// Data layer wraps unexpected database errors with this.
	ErrDataDatabase = errors.New("database error")
)
