package validation

import (
	"fmt"
	"unicode"
)

// PasswordStrength checks that a password meets minimum complexity
// requirements. It returns a ValidationError for the first unmet rule:
//   - at least minLen characters long
//   - contains at least one uppercase letter
//   - contains at least one lowercase letter
//   - contains at least one digit
//   - contains at least one special character (non-letter, non-digit)
func PasswordStrength(field, value string, minLen int) error {
	if len(value) < minLen {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s: must be at least %d characters", field, minLen),
		}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &ValidationError{
			Field:   field,
			Message: field + ": must contain at least one uppercase letter",
		}
	}
	if !hasLower {
		return &ValidationError{
			Field:   field,
			Message: field + ": must contain at least one lowercase letter",
		}
	}
	if !hasDigit {
		return &ValidationError{
			Field:   field,
			Message: field + ": must contain at least one digit",
		}
	}
	if !hasSpecial {
		return &ValidationError{
			Field:   field,
			Message: field + ": must contain at least one special character",
		}
	}

	return nil
}

// PasswordMatch checks that two password strings are identical. Use this
// to validate "confirm password" fields.
func PasswordMatch(field, password, confirm string) error {
	if password != confirm {
		return &ValidationError{
			Field:   field,
			Message: field + ": passwords do not match",
		}
	}
	return nil
}
