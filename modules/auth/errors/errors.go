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
	ErrUserDisabled      = errors.New("user account is disabled")
	ErrPasswordMismatch  = errors.New("password mismatch")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrProviderRequired  = errors.New("provider required")

	ErrInvalidOrExpiredToken = errors.New("invalid or expired token")
	ErrInvalidTokenClaims    = errors.New("invalid token claims")
	ErrInvalidTokenUser      = errors.New("invalid token user")

	// Social Account unique constraint errors
	ErrSocialEmailDoesNotExist = errors.New("social email is not linked for user")
	ErrProviderUserIDExists    = errors.New("provider user id already exists")
	ErrUserIDExists            = errors.New("user id already linked to a social account")
	ErrSocialEmailExists       = errors.New("email already linked to this provider")
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
