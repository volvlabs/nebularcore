package interfaces

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
)

// User represents the interface that all user models must implement
type User interface {
	GetID() uuid.UUID
	GetUsername() string
	GetEmail() string
	GetPhoneNumber() string
	GetPasswordHash() string
	GetRole() string
	IsActive() bool
	GetLastLoginAt() *time.Time
	GetMetadata() map[string]any
	GetPasswordResetToken() *string
	GetPasswordResetAt() *time.Time
	SetPasswordResetToken(token *string)
	SetPasswordResetAt(at *time.Time)
	SetPassword(password string)
}

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	Create(ctx context.Context, data map[string]any) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByPhoneNumber(ctx context.Context, phone string) (User, error)
	FindByIdentifier(ctx context.Context, identifier string) (User, error)
	FindByResetToken(ctx context.Context, token string) (User, error)
	UpdateLastLogin(ctx context.Context, user User) error
	ChangePassword(ctx context.Context, user User, newPasswordHash string) error
	Update(ctx context.Context, user User, updates map[string]any) error
}

// UserRepositoryFactory defines the interface for user repository factories
type UserRepositoryFactory interface {
	NewUser() User
	GetTableName() string
	GetSchema() any
}

// APICredentials represents the interface for API credentials
type APICredentials interface {
	GetID() string
	GetUserID() string
	GetAPIKey() string
	GetAPISecret() string
	GetStatus() string
	GetExpiresAt() time.Time
	GetLastUsedAt() *time.Time
}

// APIKeyRepository defines the interface for API key-related operations
type APIKeyRepository interface {
	Create(ctx context.Context, data map[string]any) (APICredentials, error)
	FindByKey(ctx context.Context, apiKey string) (APICredentials, error)
	FindByUser(ctx context.Context, userID string) ([]APICredentials, error)
	UpdateLastUsed(ctx context.Context, creds APICredentials) error
	Revoke(ctx context.Context, id string) error
}

// SocialAccount represents the interface for social login accounts
type SocialAccount interface {
	GetID() string
	GetUserID() string
	GetProvider() string
	GetProviderUserID() string
	GetMetadata() map[string]any
}

// SocialAccountRepository defines the interface for social account operations
type SocialAccountRepository interface {
	Create(ctx context.Context, data map[string]any) (SocialAccount, error)
	FindByProvider(ctx context.Context, provider, providerUserID string) (SocialAccount, error)
	FindByUser(ctx context.Context, userID string) ([]SocialAccount, error)
}

// TokenIssuer defines the interface for JWT token operations
type TokenIssuer interface {
	IssueToken(user User) (*responses.TokenResponse, error)
	ValidateToken(token string) (map[string]any, error)
	RevokeToken(token string) error
	RefreshToken(refreshToken string) (*responses.TokenResponse, error)
}

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// AuthHandler defines the interface for authentication handlers
type AuthHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
}

// PasswordHandler defines the interface for password handlers
type PasswordHandler interface {
	RegisterRoutes(router *gin.RouterGroup)
	RequestPasswordReset(c *gin.Context)
	VerifyPasswordReset(c *gin.Context)
	ChangePassword(c *gin.Context)
}

// AuthMiddleware defines the interface for authentication middleware
type AuthMiddleware interface {
	JWT() gin.HandlerFunc
	Optional() gin.HandlerFunc
	APIKey() gin.HandlerFunc
	RequireAuth() gin.HandlerFunc
	RequireRole(role string) gin.HandlerFunc
	RequirePermission(resource, action string) gin.HandlerFunc
}
