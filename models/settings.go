package models

import (
	"encoding/json"
	"errors"
	"os"

	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gopkg.in/yaml.v2"
)

type Settings struct {
	Domain                         string `yaml:"domain" json:"domain"`
	AuthTokenSecret                string `yaml:"authTokenSecret" json:"authTokenSecret"`
	AuthTokenRefreshSecret         string `yaml:"authTokenRefreshSecret" json:"authTokenRefreshSecret"`
	OtpGenerationSecret            string `yaml:"otpGenerationSecret" json:"otpGenerationSecret"`
	OtpPeriod                      uint   `yaml:"otpPeriod" json:"otpPeriod"`
	AuthTokenDuration              int64  `yaml:"authTokenDuration" json:"authTokenDuration"`
	AuthRefreshTokenExpiryDuration int64  `yaml:"authRefreshTokenExpiryDuration" json:"authRefreshTokenExpiryDuration"`

	GoogleAuth   AuthProviderConfig `yaml:"googleAuth" json:"googleAuth"`
	FacebookAuth AuthProviderConfig `yaml:"facebookAuth" json:"facebookAuth"`
	AppleAuth    AuthProviderConfig `yaml:"appleAuth" json:"appleAuth"`

	Aws         AwsConfig         `yaml:"aws" json:"aws"`
	S3          S3Config          `yaml:"s3" json:"s3"`
	CloudFront  CloudFrontConfig  `yaml:"cloudFront" json:"cloudFront"`
	EventBridge EventBridgeConfig `yaml:"eventBridge" json:"eventBridge"`

	Glcoud GcloudConfig `yaml:"gcloud" json:"gcloud"`

	EventClient types.EventClient `yaml:"eventClient" json:"eventClient"`

	AppSettings types.AppSettings `yaml:"appSettings" json:"appSettings"`
}

func NewSettings() *Settings {
	return &Settings{
		AuthTokenSecret:                "test",
		AuthTokenRefreshSecret:         "test",
		OtpGenerationSecret:            "XXXXXXXXXXXXXXXXXXXXX123A",
		OtpPeriod:                      900,
		AuthTokenDuration:              900,
		AuthRefreshTokenExpiryDuration: 172800, // 2 days.
		GoogleAuth: AuthProviderConfig{
			Enabled: false,
		},
		FacebookAuth: AuthProviderConfig{
			Enabled: false,
		},
		AppleAuth: AuthProviderConfig{
			Enabled: false,
		},
		AppSettings: map[any]any{},
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
	s.AppSettings[key] = val
}

func (s *Settings) LoadSettings(settingsFile string) error {
	f, err := os.Open(settingsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	settings := &Settings{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(settings)
	if err != nil {
		return err
	}

	return s.Merge(settings)
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
	AccessKeyID     string `yaml:"accessKeyId" json:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey" json:"secretAccessKey"`
	Region          string `yaml:"region" json:"region"`
}

type S3Config struct {
	Bucket  string `yaml:"bucket" json:"bucket"`
	Enabled bool   `yaml:"enabled" json:"enabled"`
}

type CloudFrontConfig struct {
	KeyId              string `yaml:"keyId" json:"keyId"`
	Domain             string `yaml:"domain" json:"domain"`
	PrivateKeyFilePath string `yaml:"privateKeyFilePath" json:"privateKeyFilePath"`
}

type EventBridgeConfig struct {
	EventBus string `yaml:"eventBus" json:"eventBus"`
}

type GcloudConfig struct {
	ProjectId        string `yaml:"projectId" json:"projectId"`
	CredfileLocation string `yaml:"credfileLocation" json:"credfileLocation"`

	PubSub struct {
		Topic string `yaml:"topic" json:"topic"`
	} `yaml:"pubsub" json:"pubsub"`
	Storage struct {
		Enabled          bool   `yaml:"enabled" json:"enabled"`
		Bucket           string `yaml:"bucket" json:"bucket"`
		CredfileLocation string `yaml:"credfileLocation" json:"credfileLocation"`
	} `yaml:"storage" json:"storage"`
}
