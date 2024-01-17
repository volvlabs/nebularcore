package security

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
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
