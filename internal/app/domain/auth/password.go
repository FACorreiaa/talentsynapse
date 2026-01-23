package auth

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost       = 12
	resetTokenLength = 32
)

// HashPassword creates a bcrypt hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares password against hash
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateResetToken creates a secure random token
func GenerateResetToken() (string, error) {
	bytes := make([]byte, resetTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidatePassword checks minimum requirements
func ValidatePassword(password string) bool {
	return len(password) >= 8
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	atIndex := -1
	dotAfterAt := false

	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		} else if c == '.' && atIndex != -1 {
			dotAfterAt = true
		}
	}

	return atIndex > 0 && atIndex < len(email)-1 && dotAfterAt
}
