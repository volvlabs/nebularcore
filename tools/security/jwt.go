package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

func ParseJWT(token string, verificationKey string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
	parsedToken, err := parser.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(verificationKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("unable to parse token")
}

func NewJWT(payload jwt.MapClaims, signingKey string, secondsDuration int64) (string, error) {
	if secondsDuration <= 0 {
		return "", errors.New("secondsDuration must be greater than 0")
	}

	seconds := time.Duration(secondsDuration) * time.Second
	claims := jwt.MapClaims{
		"exp": time.Now().Add(seconds).Unix(),
	}

	for k, v := range payload {
		claims[k] = v
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(signingKey))
}
