package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoneNumberValidation(t *testing.T) {
	scenarios := []struct {
		name        string
		phoneNumber string
		region      string
		expected    bool
	}{
		{
			name:        "Valid phone number",
			phoneNumber: "2348091607291",
			region:      "NG",
			expected:    true,
		},
		{
			name:        "Invalid phone number",
			phoneNumber: "23480916072",
			region:      "NG",
			expected:    false,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			validator := New()

			// Act:
			isValid := validator.ValidatePhoneNumber(scenario.phoneNumber, scenario.region)

			// Assert:
			assert.Equal(t, scenario.expected, isValid)
		})
	}
}
