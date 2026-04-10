package authentication_test

import (
	"errors"
	"github.com/volvlabs/nebularcore/services/authentication"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/models"
	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/types"
)

func TestOauth2CreateValidate(t *testing.T) {
	scenarios := []struct {
		name          string
		oauth2Request authentication.OAuth2Request
		expectedError error
	}{
		{
			name:          "should return error because of empty body",
			oauth2Request: authentication.OAuth2Request{},
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
			oauth2Request: authentication.OAuth2Request{
				Code:     "test",
				State:    "test",
				Provider: "invalid",
			},
			expectedError: errors.New("invalid provider provided"),
		},
		{
			name: "should return error because provider is not enabled",
			oauth2Request: authentication.OAuth2Request{
				Code:     "test",
				State:    "test",
				Provider: "apple",
			},
			expectedError: errors.New("provider not enabled"),
		},
		{
			name: "should return no error",
			oauth2Request: authentication.OAuth2Request{
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
			auth := authentication.New(testapp)

			otherSettings := models.NewSettings()
			otherSettings.GoogleAuth.Enabled = true
			if err := testapp.Settings().Merge(otherSettings); err != nil {
				t.Fatalf("Settings().Merge() error = %v", err)
			}

			// Act:
			err := auth.ValidateOAuth2Request(scenario.oauth2Request)

			// Assert:
			assert.Equal(t, scenario.expectedError, err)
		})
	}
}
