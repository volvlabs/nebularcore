package auth

import (
	"encoding/json"

	"github.com/volvlabs/nebularcore/tools/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const NameGoogle string = "google"

type Google struct {
	*baseProvider
}

func NewGoogleProvider() *Google {
	return &Google{&baseProvider{
		displayName: "Google",
		scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		authUrl:    google.Endpoint.AuthURL,
		tokenUrl:   google.Endpoint.TokenURL,
		userApiUrl: "https://www.googleapis.com/oauth2/v1/userinfo",
	}}
}

func (g *Google) FetchAuthUser(token *oauth2.Token) (*AuthUser, error) {
	result, err := g.sendGetUserInfoRequest(token)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(result, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}{}
	if err := json.Unmarshal(result, &extracted); err != nil {
		return nil, err
	}

	user := &AuthUser{
		Id:            extracted.Id,
		Name:          extracted.Name,
		Email:         extracted.Email,
		EmailVerified: extracted.VerifiedEmail,
		AvatarUrl:     extracted.Picture,
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		RawUser:       rawUser,
	}

	user.ExpiresAt, _ = types.ParseDateTime(token.Expiry)

	return user, nil
}
