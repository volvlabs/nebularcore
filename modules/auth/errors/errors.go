package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserExists is returned when a user already exists
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token expired")
	// ErrPermissionDenied is returned when a user doesn't have required permissions
	ErrPermissionDenied  = errors.New("permission denied")
	ErrUnsupportedMethod = errors.New("unsupported_method")
	ErrUserDisabled      = errors.New("user_disabled")
	ErrPasswordMismatch  = errors.New("password_mismatch")
	ErrInvalidPassword   = errors.New("invalid_password")
	ErrInvalidEmail      = errors.New("invalid_email")
	ErrInvalidUsername   = errors.New("invalid_username")
	ErrProviderRequired  = errors.New("provider required")
)

// AuthError represents an authentication error
type AuthError struct {
	Code    string
	Message string
}

// Error implements the error interface
func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAuthError creates a new authentication error
func NewAuthError(code string, message string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
	}
}

// IsAuthError checks if an error is an authentication error with a specific code
func IsAuthError(err error, code string) bool {
	authErr, ok := err.(*AuthError)
	return ok && authErr.Code == code
}
