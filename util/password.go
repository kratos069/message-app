package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// returns bcrypt hash of the passsword
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password")
	}

	return string(hashedPassword), nil
}

// checks if provided password is correct or not
func CheckPassword(password, hashPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashPassword), []byte(password))
}
