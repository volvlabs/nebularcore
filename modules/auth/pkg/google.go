package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidAuthorizationCode = fmt.Errorf("invalid authorization code")
	ErrInvalidIDToken           = fmt.Errorf("invalid ID token")
)

type GoogleUser struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Verified   bool   `json:"verified_email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     string `json:"locale"`
}

type GoogleSignin struct {
	config *oauth2.Config
}

func NewGoogleSignin(
	clientID string,
	clientSecret string,
	redirectURL string,
	scopes []string,
) *GoogleSignin {
	return &GoogleSignin{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		},
	}
}

func (g *GoogleSignin) Exchange(ctx context.Context, code string) (*GoogleUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		log.Err(err).Msgf("code exchange failed")
		return nil, ErrInvalidAuthorizationCode
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Err(err).Msgf("user info fetch failed")
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msgf("user info read failed")
		return nil, err
	}

	var user GoogleUser
	if err := json.Unmarshal(data, &user); err != nil {
		log.Err(err).Msgf("user info unmarshal failed")
		return nil, err
	}

	return &user, nil
}

func (g *GoogleSignin) VerifyGoogleIDToken(
	ctx context.Context,
	tokenString string,
) (*GoogleUser, error) {
	payload, err := idtoken.Validate(ctx, tokenString, g.config.ClientID)
	if err != nil {
		log.Err(err).Msgf("ID token validation failed")
		return nil, ErrInvalidIDToken
	}

	user := GoogleUser{
		ID:         payload.Subject,
		Email:      payload.Claims["email"].(string),
		Verified:   payload.Claims["email_verified"].(bool),
		Name:       payload.Claims["name"].(string),
		GivenName:  payload.Claims["given_name"].(string),
		FamilyName: payload.Claims["family_name"].(string),
		Picture:    payload.Claims["picture"].(string),
	}

	return &user, nil
}

func (g *GoogleSignin) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}
