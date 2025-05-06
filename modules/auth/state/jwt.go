package state

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
)

type JWTTokenIssuer struct {
	cfg config.JWTConfig
}

func NewJWTTokenIssuer(cfg config.JWTConfig) *JWTTokenIssuer {
	return &JWTTokenIssuer{
		cfg: cfg,
	}
}

func (t *JWTTokenIssuer) IssueToken(user interfaces.User) (*responses.TokenResponse, error) {
	claims := jwt.MapClaims{
		"sub":         user.GetID(),
		"username":    user.GetUsername(),
		"email":       user.GetEmail(),
		"phoneNumber": user.GetPhoneNumber(),
		"iat":         time.Now().Unix(),
	}

	return t.issueToken(claims, user)
}

func (t *JWTTokenIssuer) issueToken(claims jwt.MapClaims, user interfaces.User) (*responses.TokenResponse, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(t.cfg.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(t.cfg.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &responses.TokenResponse{
			AccessToken:  token,
			RefreshToken: refreshToken,
			ExpiresIn:    time.Now().Add(t.cfg.AccessTokenExpiry).Unix(),
			User:         nil,
		}, nil
	}

	return &responses.TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(t.cfg.AccessTokenExpiry).Unix(),
		User: &responses.UserResponse{
			ID:            user.GetID(),
			Username:      user.GetUsername(),
			Email:         user.GetEmail(),
			PhoneNumber:   user.GetPhoneNumber(),
			Role:          user.GetRole(),
			Metadata:      user.GetMetadata(),
			EmailVerified: user.GetEmailVerified(),
		},
	}, nil
}

func (t *JWTTokenIssuer) ValidateToken(tokenString string) (map[string]any, error) {
	return t.validateToken(tokenString, t.cfg.AccessTokenSecret)
}

func (t *JWTTokenIssuer) validateToken(tokenString, tokenSecret string) (map[string]any, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
	parsedToken, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (t *JWTTokenIssuer) RevokeToken(tokenString string) error {
	return nil
}

func (t *JWTTokenIssuer) RefreshToken(tokenString string) (*responses.TokenResponse, error) {
	claims, err := t.validateToken(tokenString, t.cfg.RefreshTokenSecret)
	if err != nil {
		return nil, err
	}

	// Update the issued at time
	claims["iat"] = time.Now().Unix()

	return t.issueToken(jwt.MapClaims(claims), nil)
}
