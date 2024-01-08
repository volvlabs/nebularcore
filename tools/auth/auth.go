package auth

import "github.com/google/uuid"

type AuthUser struct {
	Id           uuid.UUID      `json:"id"`
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	Role         string         `json:"role"`
	AvatarUrl    string         `json:"avatarUrl"`
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`
	RawUser      map[string]any `json:"rawUser"`
}
