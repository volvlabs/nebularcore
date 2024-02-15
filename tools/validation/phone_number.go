package validation

import "github.com/ttacon/libphonenumber"

func ValidatePhoneNumber(phoneNumber, region string) bool {
	parsedNumber, err := libphonenumber.Parse(phoneNumber, region)
	if err != nil {
		return false
	}

	return libphonenumber.IsValidNumber(parsedNumber)
}
