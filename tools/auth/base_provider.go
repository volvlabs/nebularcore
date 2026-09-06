package auth

import (
	"context"
	"errors"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type baseProvider struct {
	displayName  string
	clientId     string
	clientSecret string
	redirectUrl  string
	scopes       []string
	authUrl      string
	tokenUrl     string
	userApiUrl   string
}

func (b *baseProvider) DisplayName() string {
	return b.displayName
}

func (b *baseProvider) SetDisplayName(displayName string) {
	b.displayName = displayName
}

func (b *baseProvider) ClientId() string {
	return b.clientId
}

func (b *baseProvider) SetClientId(clientId string) {
	b.clientId = clientId
}

func (b *baseProvider) ClientSecret() string {
	return b.clientSecret
}

func (b *baseProvider) SetClientSecret(clientSecret string) {
	b.clientSecret = clientSecret
}

func (b *baseProvider) RedirectUrl() string {
	return b.redirectUrl
}

func (b *baseProvider) SetRedirectUrl(redirectUrl string) {
	b.redirectUrl = redirectUrl
}

func (b *baseProvider) Scopes() []string {
	return b.scopes
}

func (b *baseProvider) SetScopes(scopes []string) {
	b.scopes = scopes
}

func (b *baseProvider) AuthUrl() string {
	return b.authUrl
}

func (b *baseProvider) SetAuthUrl(authUrl string) {
	b.authUrl = authUrl
}

func (b *baseProvider) TokenUrl() string {
	return b.tokenUrl
}

func (b *baseProvider) SetTokenUrl(tokenUrl string) {
	b.tokenUrl = tokenUrl
}

func (b *baseProvider) UserApiUrl() string {
	return b.userApiUrl
}

func (b *baseProvider) SetUserApiUrl(userApiUrl string) {
	b.userApiUrl = userApiUrl
}

func (b *baseProvider) FetchToken(code string) (*oauth2.Token, error) {
	return b.oauth2Config().Exchange(context.Background(), code)
}

func (b *baseProvider) sendGetUserInfoRequest(token *oauth2.Token) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, b.userApiUrl, nil)
	if err != nil {
		return nil, err
	}

	client := b.oauth2Config().Client(context.Background(), token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("error getting OAuth2 user information")
	}

	return result, nil
}

func (b *baseProvider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     b.clientId,
		ClientSecret: b.clientSecret,
		RedirectURL:  b.redirectUrl,
		Scopes:       b.scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  b.authUrl,
			TokenURL: b.tokenUrl,
		},
	}
}
