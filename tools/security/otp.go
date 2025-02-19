package security

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
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

// Todo: allow generate to take in optional secret in other to further personalize
// token generated for each user.
func (o *Otp) Generate(secret string) (string, error) {
	key := o.GenerateUserPersonalKey(secret)
	return totp.GenerateCodeCustom(key, time.Now(), totp.ValidateOpts{
		Period: o.opts.Period,
		Digits: otp.DigitsSix,
	})
}

func (o *Otp) Validate(passcode string, secret string) bool {
	key := o.GenerateUserPersonalKey(secret)
	isValid, _ := totp.ValidateCustom(passcode, key, time.Now(), totp.ValidateOpts{
		Period: o.opts.Period,
		Digits: otp.DigitsSix,
	})
	return isValid
}

func (o *Otp) GenerateUserPersonalKey(secret string) string {
	combined := secret + "|" + o.encodedSecret
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	return base32.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hasher.Sum(nil))))
}

func GenerateUniqueOtpSecret(id uuid.UUID) string {
	hashedId := hashId(id)
	code := normalizeCode(hashedId)
	return code
}

func hashId(id uuid.UUID) string {
	userIdBytes := []byte(id.String())
	hash := sha256.Sum256(userIdBytes)
	return hex.EncodeToString(hash[:])
}

func normalizeCode(input string) string {
	normalizedCode := ""
	alphanumericCount := 0

	for _, char := range input {
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') {
			if char >= 'a' && char <= 'f' {
				char = char - 'a' + 'A'
			}

			normalizedCode += string(char)
			alphanumericCount++
		}
	}

	return normalizedCode
}
