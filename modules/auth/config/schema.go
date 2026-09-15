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
	Authorization           AuthorizationConfig          `yaml:"authorization"`
	Providers               map[string]map[string]string `yaml:"providers"`
	UserMigrationScriptPath string                       `yaml:"userMigrationScriptPath"`

	// EventsTopic is the Kafka topic auth lifecycle events (login,
	// password reset/changed, user created/updated/deleted — see
	// emitter.AuthEventData.EventType for the specific values) are
	// published to. Left empty, EmitAuthEvent is a no-op: the same
	// "not configured yet" posture other optional Kafka producers in
	// this codebase use, rather than publishing to a topic named after
	// the event type itself (which no broker provisions and fails with
	// "topic or partition does not exist").
	EventsTopic string `yaml:"eventsTopic"`
}

// AuthorizationConfig configures the module's single shared casbin
// enforcer, used both by AuthorizationManager (role/permission CRUD and
// Enforce calls made by consuming-app code) and by AuthMiddleware
// (HTTP-level RequireRole/RequirePermission gating) — see
// authorization.NewAuthorizationManager, which owns constructing the
// enforcer from this config, and middleware.NewAuthMiddleware, which now
// receives that same manager instead of building a second, independent
// enforcer (as it did previously).
type AuthorizationConfig struct {
	// Source selects where casbin policy (p) and grouping (g) rows live:
	// "database" (gorm-adapter, reading/writing through the host app's own
	// DB — the historical default behavior, kept as the zero-value default
	// for full backward compatibility) or "file" (casbin's built-in
	// file-adapter, reading/writing PolicyPath).
	Source string `yaml:"source" validate:"omitempty,oneof=database file"`

	// ModelPath, if set, loads the casbin RBAC model from this file instead
	// of the package's embedded default (rbacModelConf). Most hosts should
	// leave this empty.
	ModelPath string `yaml:"modelPath"`

	// PolicyPath is the casbin policy CSV path, required only when
	// Source == "file".
	PolicyPath string `yaml:"policyPath" validate:"required_if=Source file"`

	// MiddlewareEnabled preserves the old no-op-if-disabled behavior of
	// AuthMiddleware.RequireRole/RequirePermission — when false (the
	// default), those two middlewares are no-ops regardless of what's in
	// the enforcer, exactly like the old MiddlewareConfig.AuthorizationEnabled
	// flag. The enforcer itself is always constructed now (AuthorizationManager
	// needs it unconditionally for role/permission management), so this only
	// gates the two HTTP middlewares, not enforcer construction.
	MiddlewareEnabled bool `yaml:"middlewareEnabled"`

	// ExposeManagementAPI opts a host app into registering the generic
	// /auth/roles, /auth/roles/:name/permissions, etc. HTTP routes (see
	// handlers.NewAuthorizationHandler). Default false — this is a
	// privileged CRUD surface over roles/permissions, not every host wants
	// it registered by default.
	ExposeManagementAPI bool `yaml:"exposeManagementAPI"`
}

// JWTConfig represents JWT configuration.
//
// Algorithm controls how *access* tokens are signed: "HS256" (default, a
// shared secret only this service knows) or "RS256" (an RSA keypair, whose
// public half can be published via a JWKS endpoint so other services —
// e.g. Veda, per D6 in the mori platform plan — can verify tokens
// independently without ever holding a secret that could mint them).
// Refresh tokens always stay HS256/AccessTokenSecret-adjacent (RefreshTokenSecret):
// they're exchanged directly with this service and never need third-party
// verification, so there's no reason to widen their blast radius.
type JWTConfig struct {
	Algorithm          string        `yaml:"algorithm"`
	AccessTokenSecret  string        `yaml:"accessTokenSecret"`
	RefreshTokenSecret string        `yaml:"refreshTokenSecret"`
	PrivateKeyPEM      string        `yaml:"privateKeyPEM"`
	PublicKeyPEM       string        `yaml:"publicKeyPEM"`
	KeyID              string        `yaml:"keyId"`
	AccessTokenExpiry  time.Duration `yaml:"accessTokenExpiry"`
	RefreshTokenExpiry time.Duration `yaml:"refreshTokenExpiry"`
}

// IsRS256 reports whether access tokens should be RSA-signed.
func (c JWTConfig) IsRS256() bool {
	return c.Algorithm == "RS256"
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
	Providers map[string]SocialProviderConfig `yaml:"providers" validate:"omit_empty,dive"`
}

// SocialProviderConfig represents configuration for a social login provider
type SocialProviderConfig struct {
	ClientID     string   `yaml:"clientID" validate:"required"`
	ClientSecret string   `yaml:"clientSecret" validate:"required"`
	RedirectURL  string   `yaml:"redirectURL" validate:"required"`
	Scopes       []string `yaml:"scopes" validate:"required"`
}

// ClerkConfig represents Clerk.com configuration
type ClerkConfig struct {
	APIKey      string `yaml:"apiKey"`
	APIEndpoint string `yaml:"apiEndpoint"`
	FrontendAPI string `yaml:"frontendAPI"`
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
			Enabled:   false,
			Providers: map[string]SocialProviderConfig{},
		},
		Clerk: ClerkConfig{
			APIEndpoint: "https://api.clerk.dev/v1",
		},
		Authorization: AuthorizationConfig{
			Source: "database",
		},
	}
}

func (c *Config) Key() string {
	return "auth"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.IsRS256() {
		if c.JWT.PrivateKeyPEM == "" || c.JWT.PublicKeyPEM == "" {
			return ErrMissingJWTKeyPair
		}
	} else if c.JWT.AccessTokenSecret == "" {
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

	if c.Authorization.Source != "" && c.Authorization.Source != "database" && c.Authorization.Source != "file" {
		return ErrInvalidAuthorizationSource
	}
	if c.Authorization.Source == "file" && c.Authorization.PolicyPath == "" {
		return ErrMissingAuthorizationPolicyPath
	}

	return nil
}
