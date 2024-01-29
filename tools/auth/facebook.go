package auth

import (
	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

const NameFacebook string = "facebook"

type Facebook struct {
	*baseProvider
}

func NewFacebookProvider() *Facebook {
	return &Facebook{&baseProvider{
		displayName: "Facebook",
		scopes:      []string{"email", "public_profile"},
		authUrl:     facebook.Endpoint.AuthURL,
		tokenUrl:    facebook.Endpoint.TokenURL,
		userApiUrl:  "https://graph.facebook.com/me?fields=name,email,picture.type(large)",
	}}
}

func (f *Facebook) FetchAuthUser(token *oauth2.Token) (*AuthUser, error) {
	result, err := f.sendGetUserInfoRequest(token)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(result, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct{ Url string }
		}
	}{}
	if err := json.Unmarshal(result, &extracted); err != nil {
		return nil, err
	}

	user := &AuthUser{
		Id:            extracted.Id,
		Name:          extracted.Name,
		Email:         extracted.Email,
		EmailVerified: true,
		AvatarUrl:     extracted.Picture.Data.Url,
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		RawUser:       rawUser,
	}

	return user, nil
}
