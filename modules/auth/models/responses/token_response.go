package responses

import (
	"github.com/google/uuid"
)

// UserResponse represents user information in API responses
type UserResponse struct {
	ID            uuid.UUID      `json:"id"`
	Username      string         `json:"username"`
	Email         string         `json:"email"`
	PhoneNumber   string         `json:"phoneNumber"`
	Role          string         `json:"role"`
	EmailVerified bool           `json:"emailVerified"`
	Metadata      map[string]any `json:"metadata,omitempty"`
} // @name UserResponse

// TokenResponse represents the response after successful authentication
type TokenResponse struct {
	AccessToken  string            `json:"accessToken"`
	TokenType    string            `json:"tokenType"`
	ExpiresIn    int64             `json:"expiresIn"`
	RefreshToken string            `json:"refreshToken,omitempty"`
	Scopes       map[string]string `json:"scopes,omitempty"`
	User         *UserResponse     `json:"user,omitempty"`
} // @name TokenResponse
