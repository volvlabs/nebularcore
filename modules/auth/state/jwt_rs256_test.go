package state_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/state"
)

func generateTestKeyPair(t *testing.T) (privatePEM, publicPEM string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privatePEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	publicPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

	return privatePEM, publicPEM
}

func rs256Config(t *testing.T) config.JWTConfig {
	t.Helper()
	priv, pub := generateTestKeyPair(t)
	return config.JWTConfig{
		Algorithm:          "RS256",
		PrivateKeyPEM:      priv,
		PublicKeyPEM:       pub,
		KeyID:              "test-key-1",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenExpiry:  time.Hour,
		RefreshTokenExpiry: 24 * time.Hour,
	}
}

func TestRS256_AccessTokenVerifiableWithOnlyThePublicKey(t *testing.T) {
	cfg := rs256Config(t)
	issuer, err := state.NewJWTTokenIssuer(cfg)
	require.NoError(t, err)

	response, err := issuer.IssueToken(NewMockUser())
	require.NoError(t, err)
	require.NotEmpty(t, response.AccessToken)

	// This is the actual point of D6: a *third party* (Veda) must be able
	// to verify the token knowing only the published public key — never
	// the private key, and never any shared secret. Simulate that by
	// parsing the public key fresh from PEM, independent of the issuer.
	block, _ := pem.Decode([]byte(cfg.PublicKeyPEM))
	require.NotNil(t, block)
	pubIface, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	pub, ok := pubIface.(*rsa.PublicKey)
	require.True(t, ok)

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	parsed, err := parser.Parse(response.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return pub, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, "test-user", claims["username"])

	// kid must be present in the header so a verifier with multiple keys
	// on file knows which one to use.
	require.Equal(t, cfg.KeyID, parsed.Header["kid"])
}

func TestRS256_TokenSignedWithWrongKeyIsRejected(t *testing.T) {
	cfg := rs256Config(t)
	issuer, err := state.NewJWTTokenIssuer(cfg)
	require.NoError(t, err)

	response, err := issuer.IssueToken(NewMockUser())
	require.NoError(t, err)

	// A different keypair's public key must NOT validate this token —
	// otherwise the "asymmetric" verification would be theater.
	_, otherPub := generateTestKeyPair(t)
	block, _ := pem.Decode([]byte(otherPub))
	pubIface, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	_, err = parser.Parse(response.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return pubIface, nil
	})
	require.Error(t, err)
}

func TestJWKS_MatchesTheConfiguredPublicKey(t *testing.T) {
	cfg := rs256Config(t)
	issuer, err := state.NewJWTTokenIssuer(cfg)
	require.NoError(t, err)

	jwks, err := issuer.JWKS()
	require.NoError(t, err)

	keys, ok := jwks["keys"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, keys, 1)
	key := keys[0]
	require.Equal(t, "RSA", key["kty"])
	require.Equal(t, "RS256", key["alg"])
	require.Equal(t, cfg.KeyID, key["kid"])

	// Round-trip: rebuild an rsa.PublicKey purely from the JWKS fields (as
	// a real JWKS consumer would) and confirm it actually verifies a token
	// issued by this issuer — proving the published key is the real one,
	// not just present-but-wrong.
	nBytes := mustBase64URLDecode(t, key["n"].(string))
	eBytes := mustBase64URLDecode(t, key["e"].(string))
	rebuilt := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}

	response, err := issuer.IssueToken(NewMockUser())
	require.NoError(t, err)

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	parsed, err := parser.Parse(response.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return rebuilt, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)
}

func TestJWKS_UnavailableForHS256(t *testing.T) {
	cfg := config.JWTConfig{
		AccessTokenSecret:  "secret",
		RefreshTokenSecret: "refresh-secret",
		AccessTokenExpiry:  time.Hour,
		RefreshTokenExpiry: 24 * time.Hour,
	}
	issuer, err := state.NewJWTTokenIssuer(cfg)
	require.NoError(t, err)

	_, err = issuer.JWKS()
	require.Error(t, err, "HS256 has no public key to publish")
}

func mustBase64URLDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err)
	return b
}
