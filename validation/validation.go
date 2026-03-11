// Package validation provides input validators for common formats such as
// email addresses, UUIDs, phone numbers, and URLs. Each validator returns a
// clear, user-facing error message on failure.
package validation

import (
	"cmp"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
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

// ExactLength validates that value (after trimming) is exactly n characters.
// Useful for fixed-format fields like ISO country codes or PIN codes.
func ExactLength(field, value string, n int) error {
	v := strings.TrimSpace(value)
	if len([]rune(v)) != n {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be exactly %d characters", field, n)}
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

// Match validates that value matches the given regular expression.
// The field name and a human-readable description of the expected format
// are used in the error message.
func Match(field, value string, re *regexp.Regexp, description string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	if !re.MatchString(v) {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must match %s", field, description),
		}
	}
	return nil
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

// NoWhitespace validates that value contains no whitespace characters.
// The value is not trimmed — leading, trailing, and embedded whitespace
// all cause a validation error. Useful for usernames, API keys, and slugs.
func NoWhitespace(field, value string) error {
if value == "" {
return &ValidationError{Field: field, Message: field + " is required"}
}
if strings.ContainsAny(value, " \t\n\r") {
return &ValidationError{Field: field, Message: field + " must not contain whitespace"}
}
return nil
}

var alphaNumericRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// Alphanumeric validates that value (after trimming) contains only ASCII
// letters and digits. Useful for codes, reference IDs, and usernames.
func Alphanumeric(field, value string) error {
v := strings.TrimSpace(value)
if v == "" {
return &ValidationError{Field: field, Message: field + " is required"}
}
if !alphaNumericRe.MatchString(v) {
return &ValidationError{Field: field, Message: field + " must contain only letters and digits"}
}
return nil
}

var numericRe = regexp.MustCompile(`^[0-9]+$`)

// Numeric validates that value (after trimming) contains only ASCII digits
// (0-9). Useful for ZIP codes, phone PINs, and numeric reference codes
// that should be treated as strings rather than integers.
func Numeric(field, value string) error {
v := strings.TrimSpace(value)
if v == "" {
return &ValidationError{Field: field, Message: field + " is required"}
}
if !numericRe.MatchString(v) {
return &ValidationError{Field: field, Message: field + " must contain only digits"}
}
return nil
}

// Between validates that value is in the inclusive range [min, max].
// Works with any ordered type (int, float64, string, etc.).
func Between[V cmp.Ordered](field string, value, min, max V) error {
if value < min || value > max {
return &ValidationError{
Field:   field,
Message: fmt.Sprintf("%s must be between %v and %v", field, min, max),
}
}
return nil
}

// EachString applies a string validator to every element in values. It
// returns the first validation error encountered, using field[i] as the
// field name for error context. Returns nil if all elements pass.
func EachString(field string, values []string, validate func(string, string) error) error {
for i, v := range values {
key := fmt.Sprintf("%s[%d]", field, i)
if err := validate(key, v); err != nil {
return err
}
}
return nil
}

// Lowercase validates that value contains only lowercase letters, digits,
// and common punctuation — no uppercase letters.
func Lowercase(field, value string) error {
if value != strings.ToLower(value) {
return fmt.Errorf("%s must be lowercase", field)
}
return nil
}

// Uppercase validates that value contains only uppercase letters, digits,
// and common punctuation — no lowercase letters.
func Uppercase(field, value string) error {
if value != strings.ToUpper(value) {
return fmt.Errorf("%s must be uppercase", field)
}
return nil
}

// StartsWith validates that value starts with the given prefix.
func StartsWith(field, value, prefix string) error {
if !strings.HasPrefix(value, prefix) {
return fmt.Errorf("%s must start with %q", field, prefix)
}
return nil
}

// EndsWith validates that value ends with the given suffix.
func EndsWith(field, value, suffix string) error {
if !strings.HasSuffix(value, suffix) {
return fmt.Errorf("%s must end with %q", field, suffix)
}
return nil
}

// NotOneOf validates that value is NOT in the blocked list. Useful for
// rejecting reserved words, banned usernames, or unsafe values.
func NotOneOf(field, value string, blocked []string) error {
for _, b := range blocked {
if value == b {
return fmt.Errorf("%s must not be %q", field, value)
}
}
return nil
}

// JSON validates that s is syntactically valid JSON.
func JSON(s string) error {
if !json.Valid([]byte(s)) {
return fmt.Errorf("invalid JSON")
}
return nil
}

// Base64 validates that s is valid standard base64 (RFC 4648).
func Base64(s string) error {
_, err := base64.StdEncoding.DecodeString(s)
if err != nil {
return fmt.Errorf("invalid base64")
}
return nil
}

// IP validates that s is a valid IPv4 or IPv6 address.
func IP(s string) error {
if net.ParseIP(s) == nil {
return fmt.Errorf("invalid IP address")
}
return nil
}

// IPv4 validates that s is a valid IPv4 address.
func IPv4(s string) error {
ip := net.ParseIP(s)
if ip == nil || ip.To4() == nil {
return fmt.Errorf("invalid IPv4 address")
}
return nil
}

// CIDR validates that s is valid CIDR notation (e.g., "192.168.1.0/24").
func CIDR(s string) error {
_, _, err := net.ParseCIDR(s)
if err != nil {
return fmt.Errorf("invalid CIDR notation")
}
return nil
}
