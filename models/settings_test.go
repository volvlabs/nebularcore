package models_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/models"
)

func TestNewSettings(t *testing.T) {
	// Arrange-Act:
	settings := models.NewSettings()

	// Assert:
	assert.NotNil(t, settings)
	assert.Equal(t, "test", settings.AuthTokenSecret)
	assert.Equal(t, "XXXXXXXXXXXXXXXXXXXXX123A", settings.OtpGenerationSecret)
	assert.Equal(t, uint(900), settings.OtpPeriod)
	assert.Equal(t, int64(900), settings.AuthTokenDuration)
	assert.False(t, settings.GoogleAuth.Enabled)
	assert.False(t, settings.FacebookAuth.Enabled)
	assert.False(t, settings.AppleAuth.Enabled)
}

func TestMergeSettings(t *testing.T) {
	// Arrange:
	settings := models.NewSettings()
	otherSettings := models.NewSettings()
	rawSettingsData := []byte(`{
		"googleAuth": {
			"enabled": true,
			"clientId": "CLIENT_ID_HERE",
			"clientSecret": "CLIENT_SECRET",
			"authUrl": "https://www.authurl.com",
			"tokenUrl": "https://www.tokenurl.com",
			"displayName": "google"
		},
		"appSettings": {
			"key": "value"
		}
	}`)
	err := json.Unmarshal(rawSettingsData, otherSettings)
	if err != nil {
		t.Fatalf("error loading settings data, err: %v", err)
	}

	// Act:
	err = settings.Merge(otherSettings)
	assert.Equal(t, err, nil)
	assert.True(t, settings.GoogleAuth.Enabled)
	assert.False(t, settings.AppleAuth.Enabled)
	assert.Equal(t, settings.GoogleAuth.ClientId, "CLIENT_ID_HERE")
	assert.Equal(t, settings.GoogleAuth.ClientSecret, "CLIENT_SECRET")
	assert.Equal(t, settings.GoogleAuth.AuthUrl, "https://www.authurl.com")
	assert.Equal(t, settings.GoogleAuth.TokenUrl, "https://www.tokenurl.com")
	assert.Equal(t, settings.GoogleAuth.DisplayName, "google")
}

type MyAppSettings struct {
	Bar string `yaml:"bar" json:"bar"`
	Foo string `yaml:"foo" json:"foo"`
}

func TestLoadSettings(t *testing.T) {
	// Arrange:
	settings := models.NewSettings()
	// mySettings := &MyAppSettings{}

	// Act:
	err := settings.LoadSettings("./test_settings.yaml", &MyAppSettings{})

	// Assert:
	assert.NoError(t, err)
	appSettings := settings.AppSettings.(*MyAppSettings)
	assert.Equal(t, "foo", appSettings.Bar)
	assert.Equal(t, "baz", appSettings.Foo)
}

func TestLoadSettings__ShouldReturnErrorBecauseAppSettingsPassedIsNotAPointerType(t *testing.T) {
	// Arrange:
	settings := models.NewSettings()

	// Act:
	err := settings.LoadSettings("./test_settings.yaml", MyAppSettings{})

	// Assert:
	assert.Equal(t, models.ErrAppSettingsNotAPointer, err)
}
