package validation

import emailverifier "github.com/AfterShip/email-verifier"

func ValidateEmail(email string) (bool, error) {
	verifier := emailverifier.NewVerifier()
	result, err := verifier.Verify(email)
	if err != nil {
		return false, err
	}

	return result.Syntax.Valid, nil
}
