// Package sanitize provides input sanitization helpers for filenames and text.
package sanitize

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// DocFilename removes unsafe characters from a document filename.
func DocFilename(name string) string {
	if name == "" {
		return "unnamed"
	}
	name = filepath.Base(name)
	var clean strings.Builder
	for _, r := range name {
		if r >= 32 && r != 127 {
			clean.WriteRune(r)
		}
	}
	result := clean.String()
	if result == "" {
		result = "unnamed"
	}
	return result
}

// TruncateFilename shortens a filename to at most maxLen runes while preserving
// the file extension. If maxLen is zero or negative, returns an empty string.
// Dotfiles (e.g. ".gitignore") have no distinct extension and are returned unchanged.
func TruncateFilename(name string, maxLen int) string {
	runes := []rune(name)
	if maxLen <= 0 || len(runes) == 0 {
		return ""
	}
	if len(runes) <= maxLen {
		return name
	}
	ext := filepath.Ext(name)
	extRunes := []rune(ext)
	if len(extRunes) >= maxLen {
		return ext
	}
	base := runes[:maxLen-len(extRunes)]
	return string(base) + ext
}

// MaxLength returns s truncated to at most maxLen runes.
// If maxLen is zero or negative, returns an empty string.
func MaxLength(s string, maxLen int) string {
	runes := []rune(s)
	if maxLen <= 0 {
		return ""
	}
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// NilIfEmpty returns nil for empty strings, otherwise a pointer to s.
func NilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TrimAndNilIfEmpty trims whitespace from s and returns nil if the result is empty.
// Use this when whitespace-only strings should be treated the same as missing values.
func TrimAndNilIfEmpty(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// IsDuplicateKey reports whether err is or wraps a *pgconn.PgError with SQLSTATE 23505
// (unique-constraint violation). Plain errors.New strings are not matched.
func IsDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// SanitizeEmail trims whitespace and lowercases an email address.
// It does not validate whether the address is well-formed.
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// IsDatabaseError reports whether err is or wraps a *pgconn.PgError whose SQLSTATE code
// matches the given code. Use standard 5-character SQLSTATE codes, e.g.:
//   - "23505" – unique-constraint violation
//   - "23503" – foreign-key violation
//   - "23502" – not-null violation
//
// Plain errors created with errors.New are not matched; the error must originate from pgx.
func IsDatabaseError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// NullString returns the dereferenced value of p, or defaultVal if p is nil.
//
// Deprecated: Use Deref[string] instead.
func NullString(p *string, defaultVal string) string {
	if p == nil {
		return defaultVal
	}
	return *p
}

// NullInt64 returns the dereferenced value of p, or defaultVal if p is nil.
//
// Deprecated: Use Deref[int64] instead.
func NullInt64(p *int64, defaultVal int64) int64 {
	if p == nil {
		return defaultVal
	}
	return *p
}

// NullBool returns the dereferenced value of p, or defaultVal if p is nil.
//
// Deprecated: Use Deref[bool] instead.
func NullBool(p *bool, defaultVal bool) bool {
	if p == nil {
		return defaultVal
	}
	return *p
}

// Deref returns the dereferenced value of p, or defaultVal if p is nil.
// It is the generic equivalent of NullString, NullInt64, and NullBool.
func Deref[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// StripHTML removes all HTML tags from s, returning only the text content.
// It does not decode HTML entities (e.g. "&amp;" remains "&amp;").
func StripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Mask replaces all but the last visibleSuffix runes of s with '*'.
// Useful for masking credit card numbers, tokens, and API keys in logs.
// Returns the original string unchanged if it is shorter than or equal
// to visibleSuffix. Returns all '*' if visibleSuffix is zero or negative.
func Mask(s string, visibleSuffix int) string {
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return ""
	}
	if visibleSuffix <= 0 {
		return strings.Repeat("*", n)
	}
	if n <= visibleSuffix {
		return s
	}
	masked := strings.Repeat("*", n-visibleSuffix) + string(runes[n-visibleSuffix:])
	return masked
}

// TrimStrings returns a new slice with each element trimmed of leading and
// trailing whitespace. Empty strings after trimming are omitted.
func TrimStrings(ss []string) []string {
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
