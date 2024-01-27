package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func TestOauth2CreateValidate(t *testing.T) {
	scenarios := []struct {
		name          string
		oauth2Request services.OAuth2Request
		expectedError error
	}{
		{
			name:          "should return error because of empty body",
			oauth2Request: services.OAuth2Request{},
			expectedError: &types.RequestBodyError{
				Message: "error validating request",
				Errors: []types.FieldError{
					{
						Field:   "Code",
						Message: "Code is a required field",
					},
					{
						Field:   "State",
						Message: "State is a required field",
					},
					{
						Field:   "Provider",
						Message: "Provider is a required field",
					},
				},
			},
		},
		{

			name: "should return error because of invalid provider",
			oauth2Request: services.OAuth2Request{
				Code:     "test",
				State:    "test",
				Provider: "invalid",
			},
			expectedError: errors.New("invalid provider provided"),
		},
		{
			name: "should return error because provider is not enabled",
			oauth2Request: services.OAuth2Request{
				Code:     "test",
				State:    "test",
				Provider: "apple",
			},
			expectedError: errors.New("provider not enabled"),
		},
		{
			name: "should return no error",
			oauth2Request: services.OAuth2Request{
				Code:     "test",
				State:    "test",
				Provider: "google",
			},
			expectedError: nil,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			testapp, _ := test.NewTestApp()
			auth := services.NewAuth(testapp)

			otherSettings := models.NewSettings()
			otherSettings.GoogleAuth.Enabled = true
			testapp.Settings().Merge(otherSettings)

			// Act:
			err := auth.ValidateOAuth2Request(scenario.oauth2Request)

			// Assert:
			assert.Equal(t, scenario.expectedError, err)
		})
	}
}
