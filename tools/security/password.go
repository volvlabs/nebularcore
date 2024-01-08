package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password provided is empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func ValidatePassword(hashPassword, password string) bool {
	bytePassword := []byte(password)
	byteHashPassword := []byte(hashPassword)

	err := bcrypt.CompareHashAndPassword(byteHashPassword, bytePassword)

	return err == nil
}
