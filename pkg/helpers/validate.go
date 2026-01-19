package helpers

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a slice of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	messages := make([]string, len(ve))
	for i, e := range ve {
		messages[i] = e.Field + ": " + e.Message
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// IsEmpty checks if a string is empty or only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsEmail validates an email address
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsURL validates a URL (basic check)
func IsURL(url string) bool {
	pattern := `^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsAlphanumeric checks if a string contains only alphanumeric characters
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsNumeric checks if a string contains only numeric characters
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// MinLength checks if a string has at least n characters
func MinLength(s string, n int) bool {
	return len(strings.TrimSpace(s)) >= n
}

// MaxLength checks if a string has at most n characters
func MaxLength(s string, n int) bool {
	return len(s) <= n
}

// InRange checks if an integer is within a range (inclusive)
func InRange(n, min, max int) bool {
	return n >= min && n <= max
}

// IsStrongPassword checks if a password meets complexity requirements
// At least 8 chars, 1 uppercase, 1 lowercase, 1 digit
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// SanitizeString trims whitespace and removes control characters
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}
