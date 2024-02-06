package auth

import (
	"fmt"
	"strings"

	"gitlab.com/jideobs/nebularcore/tools/types"
	"golang.org/x/oauth2"
)

type AuthUser struct {
	Id            string         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"emailVerified"`
	AvatarUrl     string         `json:"avatarUrl"`
	AccessToken   string         `json:"accessToken"`
	RefreshToken  string         `json:"refreshToken"`
	ExpiresAt     types.DateTime `json:"expiresAt"`
	RawUser       map[string]any `json:"rawUser"`
}

func (a *AuthUser) ExtractNames() (string, string) {
	tokens := strings.Fields(a.Name)

	switch len(tokens) {
	case 0:
		return "", ""
	case 1:
		return tokens[0], ""
	case 2:
		return tokens[0], tokens[1]
	default:
		return tokens[0], tokens[len(tokens)-1]
	}
}

type Provider interface {
	DisplayName() string
	SetDisplayName(displayName string)
	ClientId() string
	SetClientId(clientId string)
	ClientSecret() string
	SetClientSecret(clientSecret string)
	RedirectUrl() string
	SetRedirectUrl(redirectUrl string)
	Scopes() []string
	SetScopes(scopes []string)
	AuthUrl() string
	SetAuthUrl(authUrl string)
	TokenUrl() string
	SetTokenUrl(tokenUrl string)
	FetchToken(code string) (*oauth2.Token, error)
	FetchAuthUser(token *oauth2.Token) (*AuthUser, error)
}

func NewProviderByName(name string) (Provider, error) {
	switch name {
	case NameApple:
		return NewAppleProvider(), nil
	case NameGoogle:
		return NewGoogleProvider(), nil
	case NameFacebook:
		return NewFacebookProvider(), nil
	default:
		return nil, fmt.Errorf("missing provider %s", name)
	}
}
