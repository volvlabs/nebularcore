package auth

import (
	"fmt"

	"gitlab.com/jideobs/nebularcore/tools/types"
	"golang.org/x/oauth2"
)

type AuthUser struct {
	Id           string         `json:"id"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	AvatarUrl    string         `json:"avatarUrl"`
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`
	ExpiresAt    types.DateTime `json:"expiresAt"`
	RawUser      map[string]any `json:"rawUser"`
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
