package security

import (
	"encoding/base32"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type OtpOptions struct {
	Secret string
	Period uint
}

type Otp struct {
	encodedSecret string
	opts          OtpOptions
}

func NewOtp(opts OtpOptions) *Otp {
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(opts.Secret))
	return &Otp{encodedSecret, opts}
}

func (o *Otp) Generate() (string, error) {
	return totp.GenerateCodeCustom(o.encodedSecret, time.Now(), totp.ValidateOpts{
		Period: o.opts.Period,
		Digits: otp.DigitsSix,
	})
}

func (o *Otp) Validate(passcode string) bool {
	isValid, _ := totp.ValidateCustom(passcode, o.encodedSecret, time.Now(), totp.ValidateOpts{
		Period: o.opts.Period,
		Digits: otp.DigitsSix,
	})
	return isValid
}
