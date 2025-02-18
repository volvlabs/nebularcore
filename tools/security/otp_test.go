package security

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestOtp(t *testing.T) {
	scenarios := []struct {
		name     string
		opts     OtpOptions
		otpValid bool
	}{
		{
			name: "should successfully generate and validate OTP",
			opts: OtpOptions{
				Period: 900,
				Secret: "secret",
			},
			otpValid: true,
		},
		{
			name: "should generate OTP but fail to validate expired OTP",
			opts: OtpOptions{
				Period: 1,
				Secret: "secret",
			},
			otpValid: false,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			otp := NewOtp(scenario.opts)
			passcode, err := otp.Generate()

			assert.Equal(t, err, nil)
			assert.NotEqual(t, passcode, "")

			time.Sleep(1 * time.Second)

			isValid := otp.Validate(passcode)

			assert.Equal(t, scenario.otpValid, isValid)
		})
	}
}

func TestGenerateUniqueOtpSecret(t *testing.T) {
	scenarios := []struct {
		name           string
		userId         uuid.UUID
		expectedLength int
		expectError    bool
	}{
		{
			name:           "should generate valid secret for regular UUID",
			userId:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			expectedLength: 64,
			expectError:    false,
		},
		{
			name:           "should generate valid secret for nil UUID",
			userId:         uuid.Nil,
			expectedLength: 64,
			expectError:    false,
		},
		{
			name:           "should generate different secrets for different UUIDs",
			userId:         uuid.MustParse("987fcdeb-51a2-43d7-9012-345678901234"),
			expectedLength: 64,
			expectError:    false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Generate the secret
			secret, err := GenerateUniqueOtpSecret(scenario.userId)

			// Check error expectation
			if scenario.expectError {
				assert.NotEqual(t, err, nil)
			} else {
				assert.Equal(t, err, nil)
			}

			// Verify secret properties
			assert.Equal(t, len(secret), scenario.expectedLength)

			// Verify secret contains only valid characters (0-9 and A-F)
			for _, char := range secret {
				isValid := (char >= '0' && char <= '9') || (char >= 'A' && char <= 'F')
				assert.Equal(t, isValid, true)
			}

			// Test idempotency - same UUID should generate same secret
			secondSecret, _ := GenerateUniqueOtpSecret(scenario.userId)
			assert.Equal(t, secret, secondSecret)

			// If testing the "different secrets" scenario, verify against first scenario
			if scenario.name == "should generate different secrets for different UUIDs" {
				firstSecret, _ := GenerateUniqueOtpSecret(scenarios[0].userId)
				assert.NotEqual(t, secret, firstSecret)
			}
		})
	}
}
