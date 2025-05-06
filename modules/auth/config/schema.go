package config

import (
	"time"
)

// Config represents the authentication module configuration
type Config struct {
	JWT                     JWTConfig                    `yaml:"jwt"`
	Backends                []string                     `yaml:"backends" validate:"required"`
	PasswordPolicy          PasswordPolicyConfig         `yaml:"passwordPolicy"`
	APIKey                  APIKeyConfig                 `yaml:"apiKey"`
	Social                  SocialConfig                 `yaml:"social"`
	Clerk                   ClerkConfig                  `yaml:"clerk"`
	MiddlewareConfig        MiddlewareConfig             `yaml:"-"`
	Providers               map[string]map[string]string `yaml:"providers"`
	UserMigrationScriptPath string                       `yaml:"userMigrationScriptPath"`
}

type MiddlewareConfig struct {
	AuthorizationEnabled bool   `yaml:"authorizationEnabled"`
	PermissionModelPath  string `yaml:"permissionModelPath"`
	PermissionPolicyPath string `yaml:"permissionPolicyPath"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	AccessTokenSecret  string        `yaml:"accessTokenSecret"`
	RefreshTokenSecret string        `yaml:"refreshTokenSecret"`
	AccessTokenExpiry  time.Duration `yaml:"accessTokenExpiry"`
	RefreshTokenExpiry time.Duration `yaml:"refreshTokenExpiry"`
}

// PasswordPolicyConfig represents password policy configuration
type PasswordPolicyConfig struct {
	MinLength      int  `yaml:"minLength"`
	RequireUpper   bool `yaml:"requireUpper"`
	RequireLower   bool `yaml:"requireLower"`
	RequireNumber  bool `yaml:"requireNumber"`
	RequireSpecial bool `yaml:"requireSpecial"`
}

// APIKeyConfig represents API key configuration
type APIKeyConfig struct {
	Length     int             `yaml:"length"`
	ExpiryTime time.Duration   `yaml:"expiryTime"`
	AllowedIPs []string        `yaml:"allowedIPs"`
	RateLimit  RateLimitConfig `yaml:"rateLimit"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Requests int           `yaml:"requests"`
	Period   time.Duration `yaml:"period"`
}

// SocialConfig represents social login configuration
type SocialConfig struct {
	Enabled   bool                            `yaml:"enabled"`
	Providers map[string]SocialProviderConfig `yaml:"providers"`
}

// SocialProviderConfig represents configuration for a social login provider
type SocialProviderConfig struct {
	ClientID     string   `yaml:"clientID"`
	ClientSecret string   `yaml:"clientSecret"`
	RedirectURL  string   `yaml:"redirectURL"`
	Scopes       []string `yaml:"scopes"`
}

// ClerkConfig represents Clerk.com configuration
type ClerkConfig struct {
	APIKey      string `yaml:"apiKey"`
	APIEndpoint string `yaml:"apiEndpoint"`
	FrontendAPI string `yaml:"frontendAPI"`
}

func (c *Config) ConfigKey() string {
	return "auth"
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		JWT: JWTConfig{
			AccessTokenExpiry:  24 * time.Hour,
			RefreshTokenExpiry: 720 * time.Hour, // 30 days
		},
		PasswordPolicy: PasswordPolicyConfig{
			MinLength:      8,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumber:  true,
			RequireSpecial: true,
		},
		APIKey: APIKeyConfig{
			Length:     32,
			ExpiryTime: 8760 * time.Hour, // 1 year
			RateLimit: RateLimitConfig{
				Enabled:  true,
				Requests: 1000,
				Period:   time.Hour,
			},
		},
		Social: SocialConfig{
			Enabled: false,
			Providers: map[string]SocialProviderConfig{
				"google": {
					Scopes: []string{
						"https://www.googleapis.com/auth/userinfo.email",
						"https://www.googleapis.com/auth/userinfo.profile",
					},
				},
				"github": {
					Scopes: []string{
						"user:email",
						"read:user",
					},
				},
				"facebook": {
					Scopes: []string{
						"email",
						"public_profile",
					},
				},
			},
		},
		Clerk: ClerkConfig{
			APIEndpoint: "https://api.clerk.dev/v1",
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.AccessTokenSecret == "" {
		return ErrMissingJWTSecret
	}

	if c.JWT.RefreshTokenSecret == "" {
		return ErrMissingRefreshTokenSecret
	}

	if c.JWT.AccessTokenExpiry <= 0 {
		return ErrInvalidJWTExpiry
	}

	if c.JWT.RefreshTokenExpiry <= 0 {
		return ErrInvalidRefreshExpiry
	}

	if c.PasswordPolicy.MinLength < 8 {
		return ErrInvalidPasswordLength
	}

	if c.APIKey.Length < 16 {
		return ErrInvalidAPIKeyLength
	}

	if c.Social.Enabled {
		for provider, cfg := range c.Social.Providers {
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return ErrInvalidSocialConfig{Provider: provider}
			}
		}
	}

	if c.Clerk.APIKey != "" && c.Clerk.APIEndpoint == "" {
		return ErrInvalidClerkConfig
	}

	return nil
}
