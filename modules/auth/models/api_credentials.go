package models

import "time"

// APICredentials represents API key credentials
type APICredentials struct {
	APIKey    string    `json:"api_key"`
	APISecret string    `json:"api_secret"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
