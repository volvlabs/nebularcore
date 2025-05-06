package config

import "fmt"

// Configuration errors
var (
	ErrMissingJWTSecret          = fmt.Errorf("JWT secret is required")
	ErrMissingRefreshTokenSecret = fmt.Errorf("JWT refresh token secret is required")
	ErrInvalidJWTExpiry          = fmt.Errorf("invalid JWT expiry time")
	ErrInvalidRefreshExpiry      = fmt.Errorf("invalid refresh token expiry time")
	ErrInvalidPasswordLength     = fmt.Errorf("minimum password length must be at least 8 characters")
	ErrInvalidAPIKeyLength       = fmt.Errorf("API key length must be at least 16 characters")
	ErrInvalidClerkConfig        = fmt.Errorf("clerk API endpoint is required when API key is provided")
)

// ErrInvalidSocialConfig represents an error with social provider configuration
type ErrInvalidSocialConfig struct {
	Provider string
}

func (e ErrInvalidSocialConfig) Error() string {
	return fmt.Sprintf("invalid social provider configuration for %s: client ID and secret are required", e.Provider)
}
