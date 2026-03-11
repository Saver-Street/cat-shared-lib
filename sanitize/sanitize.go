// Package sanitize provides input sanitization helpers for filenames and text.
package sanitize

import (
	"errors"
	"html"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

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

var multiHyphen = regexp.MustCompile(`-{2,}`)

// Slugify converts s into a URL-friendly slug. It lowercases the string,
// replaces non-alphanumeric characters with hyphens, collapses consecutive
// hyphens, and trims leading/trailing hyphens.
func Slugify(s string) string {
s = strings.TrimSpace(s)
s = strings.ToLower(s)

var b strings.Builder
b.Grow(len(s))
for _, r := range s {
if unicode.IsLetter(r) || unicode.IsDigit(r) {
b.WriteRune(r)
} else {
b.WriteByte('-')
}
}
slug := multiHyphen.ReplaceAllString(b.String(), "-")
return strings.Trim(slug, "-")
}

// EscapeHTML escapes special HTML characters (<, >, &, ', ") so the string
// can be safely embedded in HTML content without risk of injection.
func EscapeHTML(s string) string {
return html.EscapeString(s)
}

// Truncate shortens s to at most maxLen runes. If the string exceeds maxLen,
// it is trimmed and "…" is appended (the ellipsis counts toward maxLen).
// Useful for display names, log messages, and UI labels.
func Truncate(s string, maxLen int) string {
runes := []rune(s)
if len(runes) <= maxLen {
return s
}
if maxLen <= 1 {
return "…"
}
return string(runes[:maxLen-1]) + "…"
}

// RemoveNonPrintable strips control characters and non-printable runes from s,
// keeping only printable Unicode characters and ASCII whitespace (space, tab,
// newline). Useful for cleaning user input before storage or display.
func RemoveNonPrintable(s string) string {
var b strings.Builder
b.Grow(len(s))
for _, r := range s {
if unicode.IsPrint(r) || r == '\n' || r == '\t' {
b.WriteRune(r)
}
}
return b.String()
}

var multiSpace = regexp.MustCompile(`\s+`)

// NormalizeWhitespace collapses all consecutive whitespace (spaces, tabs,
// newlines) into a single space and trims leading/trailing whitespace.
// Useful for cleaning user-entered names, titles, and search queries.
func NormalizeWhitespace(s string) string {
return strings.TrimSpace(multiSpace.ReplaceAllString(s, " "))
}

// CamelToSnake converts a camelCase or PascalCase string to snake_case.
// Consecutive uppercase letters (e.g., "HTTPClient") are treated as a
// single acronym ("http_client"). Empty input returns "".
func CamelToSnake(s string) string {
if s == "" {
return ""
}
var b strings.Builder
runes := []rune(s)
for i, r := range runes {
if unicode.IsUpper(r) {
if i > 0 {
prev := runes[i-1]
// Insert underscore before an uppercase letter when the
// previous character is lowercase, OR when the previous
// character is uppercase and the next is lowercase (end
// of an acronym like "HTTP" in "HTTPClient").
if unicode.IsLower(prev) ||
(unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
b.WriteRune('_')
}
}
b.WriteRune(unicode.ToLower(r))
} else {
b.WriteRune(r)
}
}
return b.String()
}

// SnakeToCamel converts a snake_case string to camelCase. The first
// character remains lowercase. Empty segments (from consecutive
// underscores) are skipped.
func SnakeToCamel(s string) string {
parts := strings.Split(s, "_")
var b strings.Builder
for i, p := range parts {
if p == "" {
continue
}
if i == 0 || b.Len() == 0 {
b.WriteString(strings.ToLower(p))
} else {
runes := []rune(strings.ToLower(p))
runes[0] = unicode.ToUpper(runes[0])
b.WriteString(string(runes))
}
}
return b.String()
}
