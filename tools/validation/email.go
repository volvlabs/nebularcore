package validation

import (
	"strings"

	emailverifier "github.com/AfterShip/email-verifier"
)

func ValidateEmail(email string) (bool, error) {
	verifier := emailverifier.NewVerifier()
	result, err := verifier.Verify(email)
	if err != nil {
		if strings.Contains(err.Error(), "Mail server does not exist") {
			return false, nil
		}
		return false, err
	}

	return result.Syntax.Valid, nil
}
