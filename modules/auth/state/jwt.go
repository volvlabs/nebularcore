package state

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models/responses"
)

type JWTTokenIssuer struct {
	cfg        config.JWTConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	roles      interfaces.RoleLister
	orgs       interfaces.OrgLister
}

// SetRoleLister wires in an optional roles source (see
// interfaces.RoleListerSetter's doc comment for why this is a
// post-construction setter rather than a constructor param). Safe to call
// with nil to disable — IssueToken simply omits the "roles" claim then,
// same as if this were never called.
func (t *JWTTokenIssuer) SetRoleLister(rl interfaces.RoleLister) {
	t.roles = rl
}

// SetOrgLister wires in an optional org source, mirroring SetRoleLister
// (see interfaces.OrgListerSetter's doc comment). Safe to call with nil to
// disable — IssueToken simply omits the "org_id" claim then.
func (t *JWTTokenIssuer) SetOrgLister(ol interfaces.OrgLister) {
	t.orgs = ol
}

var _ interfaces.RoleListerSetter = (*JWTTokenIssuer)(nil)
var _ interfaces.OrgListerSetter = (*JWTTokenIssuer)(nil)

// NewJWTTokenIssuer constructs a token issuer. When cfg.Algorithm is
// "RS256", the configured PEM key pair is parsed up front so a
// misconfiguration fails at startup rather than on the first login.
func NewJWTTokenIssuer(cfg config.JWTConfig) (*JWTTokenIssuer, error) {
	issuer := &JWTTokenIssuer{cfg: cfg}

	if cfg.IsRS256() {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parsing JWT private key: %w", err)
		}
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cfg.PublicKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("parsing JWT public key: %w", err)
		}
		issuer.privateKey = privateKey
		issuer.publicKey = publicKey
	}

	return issuer, nil
}

func (t *JWTTokenIssuer) IssueToken(user interfaces.User) (*responses.TokenResponse, error) {
	claims := jwt.MapClaims{
		"sub":         user.GetID(),
		"username":    user.GetUsername(),
		"email":       user.GetEmail(),
		"phoneNumber": user.GetPhoneNumber(),
		"iat":         time.Now().Unix(),
	}

	// Best-effort: a roles-lookup failure must not block login — it just
	// means the token comes back without a "roles" claim, same as if no
	// RoleLister were configured at all. IssueToken has no request-scoped
	// context of its own (see TokenIssuer's interface), so this uses a
	// background one for the lookup.
	if t.roles != nil {
		roles, err := t.roles.GetUserRoles(context.Background(), user.GetID().String())
		if err != nil {
			log.Err(err).Str("user", user.GetID().String()).Msg("auth: failed to look up roles for JWT claim")
		} else {
			claims["roles"] = roles
		}
	}

	// Best-effort, same posture as roles above: an org lookup failure must
	// not block login, it just means the token comes back without an
	// "org_id" claim.
	if t.orgs != nil {
		orgID, err := t.orgs.GetUserOrgID(context.Background(), user.GetID().String())
		if err != nil {
			log.Err(err).Str("user", user.GetID().String()).Msg("auth: failed to look up org for JWT claim")
		} else if orgID != "" {
			claims["org_id"] = orgID
		}
	}

	return t.issueToken(claims, user)
}

func (t *JWTTokenIssuer) issueToken(claims jwt.MapClaims, user interfaces.User) (*responses.TokenResponse, error) {
	accessToken, err := t.signAccessToken(claims)
	if err != nil {
		return nil, err
	}

	// Refresh tokens are exchanged directly with this service and never
	// need independent third-party verification, so they always stay
	// HS256 regardless of the access token's algorithm.
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(t.cfg.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &responses.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    time.Now().Add(t.cfg.AccessTokenExpiry).Unix(),
			User:         nil,
		}, nil
	}

	return &responses.TokenResponse{
		AccessToken:  accessToken,
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

func (t *JWTTokenIssuer) signAccessToken(claims jwt.MapClaims) (string, error) {
	if t.cfg.IsRS256() {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		if t.cfg.KeyID != "" {
			token.Header["kid"] = t.cfg.KeyID
		}
		return token.SignedString(t.privateKey)
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(t.cfg.AccessTokenSecret))
}

// ValidateToken validates an access token using whichever algorithm is
// configured.
func (t *JWTTokenIssuer) ValidateToken(tokenString string) (map[string]any, error) {
	var (
		validMethod string
		key         interface{}
	)
	if t.cfg.IsRS256() {
		validMethod = "RS256"
		key = t.publicKey
	} else {
		validMethod = "HS256"
		key = []byte(t.cfg.AccessTokenSecret)
	}

	return parseAndValidate(tokenString, validMethod, key)
}

func (t *JWTTokenIssuer) validateToken(tokenString, tokenSecret string) (map[string]any, error) {
	return parseAndValidate(tokenString, "HS256", []byte(tokenSecret))
}

func parseAndValidate(tokenString, validMethod string, key interface{}) (map[string]any, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{validMethod}))
	parsedToken, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return key, nil
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

// JWKS returns the access-token public key as a standard JSON Web Key Set
// (RFC 7517), for publishing at a well-known endpoint so other services
// (e.g. Veda, per D6 in the mori platform plan) can verify access tokens
// without ever holding a secret capable of minting them. Only meaningful
// when configured for RS256 — HS256 has no public half to publish.
func (t *JWTTokenIssuer) JWKS() (map[string]any, error) {
	if !t.cfg.IsRS256() {
		return nil, fmt.Errorf("JWKS is only available when JWT.algorithm is RS256")
	}

	return map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": t.cfg.KeyID,
				"n":   base64.RawURLEncoding.EncodeToString(t.publicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(t.publicKey.E)).Bytes()),
			},
		},
	}, nil
}
