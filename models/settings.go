package models

import (
	"encoding/json"
	"errors"

	"gitlab.com/jideobs/nebularcore/tools/auth"
)

type Settings struct {
	AuthTokenSecret     string `json:"authTokenSecret"`
	OtpGenerationSecret string `json:"otpGenerationSecret"`
	OtpPeriod           uint   `json:"otpPeriod"`
	AuthTokenDuration   int64  `json:"authTokenDuration"`

	GoogleAuth   AuthProviderConfig `json:"googleAuth"`
	FacebookAuth AuthProviderConfig `json:"facebookAuth"`
	AppleAuth    AuthProviderConfig `json:"appleAuth"`

	Aws AwsConfig `json:"aws"`
	S3  S3Config  `json:"s3"`

	OtherSettings map[string]any `json:"otherSettings"`
}

func NewSettings() *Settings {
	return &Settings{
		AuthTokenSecret:     "test",
		OtpGenerationSecret: "XXXXXXXXXXXXXXXXXXXXX123A",
		OtpPeriod:           900,
		AuthTokenDuration:   900,
		GoogleAuth: AuthProviderConfig{
			Enabled: false,
		},
		FacebookAuth: AuthProviderConfig{
			Enabled: false,
		},
		AppleAuth: AuthProviderConfig{
			Enabled: false,
		},
		OtherSettings: map[string]any{},
	}
}

func (s *Settings) Merge(other *Settings) error {
	bytes, err := json.Marshal(other)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}

func (s *Settings) NamedAuthProviderConfig(providerName string) (AuthProviderConfig, bool) {
	providerConfigs := map[string]AuthProviderConfig{
		auth.NameApple:    s.AppleAuth,
		auth.NameFacebook: s.FacebookAuth,
		auth.NameGoogle:   s.GoogleAuth,
	}
	config, ok := providerConfigs[providerName]
	return config, ok
}

func (s *Settings) AddOtherSetting(key string, val any) {
	s.OtherSettings[key] = val
}

type AuthProviderConfig struct {
	Enabled      bool   `json:"enabled"`
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AuthUrl      string `json:"authUrl"`
	TokenUrl     string `json:"tokenUrl"`
	DisplayName  string `json:"displayName"`
}

func (a AuthProviderConfig) SetupProvider(provider auth.Provider) error {
	if !a.Enabled {
		return errors.New("the provider is not enabled")
	}

	if a.ClientId != "" {
		provider.SetClientId(a.ClientId)
	}

	if a.ClientSecret != "" {
		provider.SetClientSecret(a.ClientSecret)
	}

	if a.AuthUrl != "" {
		provider.SetAuthUrl(a.AuthUrl)
	}

	if a.TokenUrl != "" {
		provider.SetTokenUrl(a.TokenUrl)
	}

	if a.DisplayName != "" {
		provider.SetDisplayName(a.DisplayName)
	}

	return nil
}

type AwsConfig struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

type S3Config struct {
	Bucket  string `json:"bucket"`
	Enabled bool   `json:"enabled"`
}
