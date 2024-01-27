package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cast"
	"gitlab.com/jideobs/nebularcore/tools/httpclient"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"golang.org/x/oauth2"
)

const NameApple string = "apple"

type Apple struct {
	*baseProvider

	publicKeyUrl string
}

func NewAppleProvider() *Apple {
	return &Apple{
		baseProvider: &baseProvider{
			displayName: "Apple",
			scopes:      nil,
			authUrl:     "https://appleid.apple.com/auth/authorize",
			tokenUrl:    "https://appleid.apple.com/auth/token",
		},
		publicKeyUrl: "https://appleid.apple.com/auth/keys",
	}
}

func (a *Apple) FetchAuthUser(token *oauth2.Token) (*AuthUser, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("id_token not found")
	}

	data, err := a.parseToken(idToken)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Id            string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	user := &AuthUser{
		Id:           extracted.Id,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		RawUser:      rawUser,
	}

	user.ExpiresAt, _ = types.ParseDateTime(token.Expiry)
	if cast.ToBool(extracted.EmailVerified) {
		user.Email = extracted.Email
	}

	return user, nil
}

func (a *Apple) parseToken(idToken string) ([]byte, error) {
	claims := jwt.MapClaims{}
	t, _, err := jwt.NewParser().ParseUnverified(idToken, claims)
	if err != nil {
		return nil, err
	}

	if !claims.VerifyIssuer("https://appleid.apple.com", true) {
		return nil, errors.New("iss must be https://appleid.apple.com")
	}

	if !claims.VerifyAudience(a.clientId, true) {
		return nil, errors.New("aud must be " + a.clientId)
	}

	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.New("kid not found")
	}

	jwk, err := a.fetchJWK(kid)
	if err != nil {
		return nil, err
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwk.Alg}))

	exponent, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(jwk.E, "="))
	if err != nil {
		return nil, err
	}

	modulus, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(jwk.N, "="))
	if err != nil {
		return nil, err
	}

	publicKey := &rsa.PublicKey{
		E: int(big.NewInt(0).SetBytes(exponent).Uint64()),
		N: big.NewInt(0).SetBytes(modulus),
	}

	parsedToken, err := parser.Parse(idToken, func(token *jwt.Token) (any, error) {
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok = parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("the parsed id_token is invalid")
	}

	return json.Marshal(claims)
}

type jwk struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

func (a *Apple) fetchJWK(kid string) (*jwk, error) {
	httpClient := httpclient.NewHTTPClient(a.publicKeyUrl, "")

	jwks := struct {
		Keys []*jwk
	}{}

	resp, err := httpClient.
		ToJson(&jwks).
		Get("")

	if err != nil {
		return nil, errors.New("error occurred trying to get public key")
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("error response received trying to get public key")
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return key, nil
		}
	}

	return nil, fmt.Errorf("jwk with kid %s not found", kid)
}
