package errutil

import "fmt"

// AppError is a structured error with an error code and optional details.
type AppError struct {
	Code    string      // error code, e.g. ERR_INVALID_INPUT, ERR_SCHEMA_VALIDATION_FAILED
	Message string      // human readable message
	Details interface{} // optional additional context
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil error>"
	}
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Predefined error codes
const (
	ErrCodeInvalidInput     = "ERR_INVALID_INPUT"
	ErrCodeSchemaValidation = "ERR_SCHEMA_VALIDATION_FAILED"
	ErrCodeNotFound         = "ERR_NOT_FOUND"
	ErrCodeInternal         = "ERR_INTERNAL"
)

// NewAppError creates a new AppError
func NewAppError(code, message string, details interface{}) error {
	return &AppError{Code: code, Message: message, Details: details}
}
