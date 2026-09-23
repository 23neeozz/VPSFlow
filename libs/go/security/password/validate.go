package password

import (
	"strings"
	"unicode"

	"github.com/bosscloud/bosscloud/libs/go/errors"
)

// Validate enforces BossCloud password policy.
func Validate(password string) error {
	if len(password) < 12 {
		return errors.New(errors.CodeValidation, "password must be at least 12 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:,.<>?", r):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New(errors.CodeValidation, "password must include uppercase, lowercase, digit, and special character")
	}

	return nil
}
