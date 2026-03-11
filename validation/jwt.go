package validation

import (
	"encoding/base64"
	"strings"
)

// JWTFormat validates that value looks like a well-formed JSON Web Token
// (three base64url-encoded segments separated by dots). It does NOT verify
// signatures or claims.
func JWTFormat(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "JWT is required"}
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return &ValidationError{Field: field, Message: "JWT must have three dot-separated segments"}
	}
	for i, p := range parts {
		if p == "" {
			return &ValidationError{
				Field:   field,
				Message: "JWT segment " + segmentName(i) + " must not be empty",
			}
		}
		if _, err := base64.RawURLEncoding.DecodeString(p); err != nil {
			return &ValidationError{
				Field:   field,
				Message: "JWT segment " + segmentName(i) + " is not valid base64url",
			}
		}
	}
	return nil
}

func segmentName(i int) string {
	switch i {
	case 0:
		return "header"
	case 1:
		return "payload"
	default:
		return "signature"
	}
}

// SemVer validates that value is a valid semantic version (e.g. "1.2.3"
// or "1.2.3-beta+build"). The "v" prefix is optional.
func SemVer(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "semantic version is required"}
	}
	s := strings.TrimPrefix(value, "v")
	// Split off build metadata
	s, _ = cutLast(s, "+")
	// Split off pre-release
	core, _, _ := strings.Cut(s, "-")

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return &ValidationError{Field: field, Message: "semantic version must have three numeric parts (MAJOR.MINOR.PATCH)"}
	}
	for _, p := range parts {
		if p == "" || !isDigits(p) || (len(p) > 1 && p[0] == '0') {
			return &ValidationError{Field: field, Message: "semantic version parts must be non-negative integers without leading zeros"}
		}
	}
	return nil
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func cutLast(s, sep string) (before string, after string) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+len(sep):]
}
