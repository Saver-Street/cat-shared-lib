// Package validate provides common input validators for email addresses,
// UUIDs, phone numbers, and URLs.
package validate

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// emailRe matches a simplified email pattern: local@domain.tld
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// uuidRe matches standard UUID v1–v5 (8-4-4-4-12 hex, case-insensitive).
	uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// phoneRe matches E.164 format: optional +, 7-15 digits.
	phoneRe = regexp.MustCompile(`^\+?[0-9]{7,15}$`)
)

// Email reports whether s is a syntactically valid email address.
func Email(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 254 {
		return false
	}
	return emailRe.MatchString(s)
}

// UUID reports whether s is a valid UUID (v1–v5 format).
func UUID(s string) bool {
	return uuidRe.MatchString(strings.TrimSpace(s))
}

// Phone reports whether s is a valid phone number in E.164 format.
// Accepts optional leading +, followed by 7-15 digits.
func Phone(s string) bool {
	s = strings.TrimSpace(s)
	// Strip common formatting characters before validation.
	s = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(s)
	return phoneRe.MatchString(s)
}

// URL reports whether s is a valid absolute HTTP or HTTPS URL.
func URL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}
