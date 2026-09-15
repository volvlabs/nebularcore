package types

import "errors"

var ErrRecordNotFound = errors.New("record not found")

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorType represents the category of error
type ErrorType int

const (
	ErrorTypeSystem ErrorType = iota
	ErrorTypeUser
	ErrorTypeValidation
)

// AppError represents a unified error type with additional context
type AppError struct {
	Type    ErrorType
	Message string
	Errors  []FieldError // Optional validation errors
}

func (e *AppError) Error() string {
	return e.Message
}

// NewUserError creates an error that represents a user mistake
func NewUserError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUser,
		Message: message,
	}
}

// NewValidationError creates an error for request validation failures
func NewValidationError(message string, errors []FieldError) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Errors:  errors,
	}
}

// NewSystemError creates an error for internal system failures
func NewSystemError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeSystem,
		Message: message,
	}
}

func ErrIsUserError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrorTypeUser || appErr.Type == ErrorTypeValidation
	}
	return false
}
