package backends

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/jideobs/nebularcore/modules/auth/config"
)

// SocialProviderClient defines the interface for social provider API clients
type SocialProviderClient interface {
	GetUserInfo(accessToken string) (*SocialUserInfo, error)
}

// SocialUserInfo represents common user info from social providers
type SocialUserInfo struct {
	ID    string
	Email string
	Name  string
	Raw   map[string]interface{}
}

// GoogleClient implements Google OAuth2 authentication
type GoogleClient struct {
	config config.SocialProviderConfig
}

func NewGoogleClient(config config.SocialProviderConfig) *GoogleClient {
	return &GoogleClient{config: config}
}

func (c *GoogleClient) GetUserInfo(accessToken string) (*SocialUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &SocialUserInfo{
		ID:    raw["id"].(string),
		Email: raw["email"].(string),
		Name:  raw["name"].(string),
		Raw:   raw,
	}, nil
}

// GithubClient implements Github OAuth authentication
type GithubClient struct {
	config config.SocialProviderConfig
}

func NewGithubClient(config config.SocialProviderConfig) *GithubClient {
	return &GithubClient{config: config}
}

func (c *GithubClient) GetUserInfo(accessToken string) (*SocialUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &SocialUserInfo{
		ID:    raw["id"].(string),
		Email: raw["email"].(string),
		Name:  raw["name"].(string),
		Raw:   raw,
	}, nil
}

// FacebookClient implements Facebook OAuth authentication
type FacebookClient struct {
	config config.SocialProviderConfig
}

func NewFacebookClient(config config.SocialProviderConfig) *FacebookClient {
	return &FacebookClient{config: config}
}

func (c *FacebookClient) GetUserInfo(accessToken string) (*SocialUserInfo, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,email,name&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &SocialUserInfo{
		ID:    raw["id"].(string),
		Email: raw["email"].(string),
		Name:  raw["name"].(string),
		Raw:   raw,
	}, nil
}
