// Package validation provides input validators for common formats such as
// email addresses, UUIDs, phone numbers, and URLs. Each validator returns a
// clear, user-facing error message on failure.
package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// emailRe matches the vast majority of valid email addresses.
// It intentionally rejects edge-case addresses (quoted local-parts, IP
// literals) that are rarely seen in practice.
var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// uuidRe matches standard UUID v1–v5 (8-4-4-4-12 hex digits).
var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// phoneRe matches E.164 phone numbers: optional +, 7–15 digits, optional
// dashes/spaces/dots between groups.
var phoneRe = regexp.MustCompile(`^\+?[0-9][0-9\-\.\s]{5,14}[0-9]$`)

// Compile-time interface compliance check.
var _ error = (*ValidationError)(nil)

// ValidationError holds a field name and a human-readable error message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Email validates that value is a well-formed email address.
func Email(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: "email is required"}
	}
	if len(v) > 254 {
		return &ValidationError{Field: field, Message: "email exceeds maximum length of 254 characters"}
	}
	if !emailRe.MatchString(v) {
		return &ValidationError{Field: field, Message: "invalid email format"}
	}
	return nil
}

// UUID validates that value is a valid UUID (v1–v5, RFC 4122 format).
func UUID(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: "UUID is required"}
	}
	if !uuidRe.MatchString(v) {
		return &ValidationError{Field: field, Message: "invalid UUID format"}
	}
	return nil
}

// Phone validates that value looks like a valid phone number.
// Accepts E.164 and common formats with dashes, dots, or spaces.
func Phone(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: "phone number is required"}
	}
	// Strip common formatting characters for length check
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, v)
	if len(digits) < 7 || len(digits) > 15 {
		return &ValidationError{Field: field, Message: "phone number must contain 7-15 digits"}
	}
	if !phoneRe.MatchString(v) {
		return &ValidationError{Field: field, Message: "invalid phone number format"}
	}
	return nil
}

// URL validates that value is a well-formed absolute URL with http or https scheme.
func URL(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: "URL is required"}
	}
	u, err := url.Parse(v)
	if err != nil {
		return &ValidationError{Field: field, Message: "invalid URL format"}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return &ValidationError{Field: field, Message: "URL must use http or https scheme"}
	}
	if u.Host == "" {
		return &ValidationError{Field: field, Message: "URL must include a host"}
	}
	return nil
}

// Required validates that value is non-empty after trimming whitespace.
func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	return nil
}

// MinLength validates that value (after trimming) has at least min runes.
func MinLength(field, value string, min int) error {
	v := strings.TrimSpace(value)
	if len([]rune(v)) < min {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be at least %d characters", field, min)}
	}
	return nil
}

// MaxLength validates that value (after trimming) has at most max runes.
func MaxLength(field, value string, max int) error {
	v := strings.TrimSpace(value)
	if len([]rune(v)) > max {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be at most %d characters", field, max)}
	}
	return nil
}

// OneOf validates that value is one of the allowed values.
func OneOf(field, value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")),
	}
}

// Collect runs multiple validation functions and returns all errors.
// Returns nil if all validations pass.
func Collect(errs ...error) []error {
	var result []error
	for _, err := range errs {
		if err != nil {
			result = append(result, err)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// slugRe matches URL-friendly slugs: lowercase letters, digits, and hyphens.
// Must start and end with an alphanumeric character.
var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Slug validates that value is a URL-friendly slug: lowercase letters, digits,
// and hyphens only, starting and ending with an alphanumeric character.
func Slug(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	if !slugRe.MatchString(v) {
		return &ValidationError{Field: field, Message: field + " must contain only lowercase letters, digits, and hyphens"}
	}
	return nil
}
