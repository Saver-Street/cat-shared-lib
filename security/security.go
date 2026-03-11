// Package security provides input validation and PII redaction helpers.
package security

import (
	"net/url"
	"regexp"
	"strings"
)

// suspiciousPatterns is the list of compiled regexes used by ContainsSuspiciousInput
// to detect SQL injection and HTML/JS injection attempts.
var suspiciousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)DROP\s+TABLE`),
	regexp.MustCompile(`(?i)SELECT\s+\*\s+FROM`),
	regexp.MustCompile(`(?i)UNION\s+SELECT`),
	regexp.MustCompile(`(?i)INSERT\s+INTO`),
	regexp.MustCompile(`(?i)DELETE\s+FROM`),
	regexp.MustCompile(`(?i)UPDATE\s+\w+\s+SET`),
	regexp.MustCompile(`(?i)<script`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
	regexp.MustCompile(`(?i)<iframe`),
	regexp.MustCompile(`(?i)<object`),
	regexp.MustCompile(`(?i)<embed`),
	regexp.MustCompile(`(?i)<svg[^>]*on`),
	regexp.MustCompile(`(?i)data\s*:\s*text/html`),
}

// ContainsSuspiciousInput returns true if the value matches known SQL injection or XSS patterns.
func ContainsSuspiciousInput(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" {
		return false
	}
	for _, p := range suspiciousPatterns {
		if p.MatchString(v) {
			return true
		}
	}
	return false
}

// piiFields is the set of JSON field names that RedactPII replaces with "[REDACTED]".
// Matching is case-insensitive (both the literal key and its lowercased form are checked).
var piiFields = map[string]bool{
	"email": true, "phone": true, "address": true, "ssn": true,
	"password": true, "resume": true, "socialSecurityNumber": true,
	"phoneNumber": true, "emailAddress": true, "streetAddress": true,
	"zipCode": true, "postalCode": true, "dateOfBirth": true,
}

// emailRe, phoneRe, and ssnRe are used by redactString to replace PII patterns
// found inside string values during RedactPII processing.
var (
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRe = regexp.MustCompile(`(\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	ssnRe   = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)
)

// maxRedactDepth is the maximum nesting depth for RedactPII to prevent stack overflow
// from pathologically nested payloads.
const maxRedactDepth = 20

// RedactPII performs server-side PII scrubbing on audit payloads.
// Nested maps are processed recursively up to maxRedactDepth levels.
func RedactPII(data map[string]any) map[string]any {
	return redactMap(data, 0)
}

func redactMap(data map[string]any, depth int) map[string]any {
	result := make(map[string]any, len(data))
	for k, v := range data {
		if piiFields[k] || piiFields[strings.ToLower(k)] {
			result[k] = "[REDACTED]"
			continue
		}
		result[k] = redactValue(v, depth)
	}
	return result
}

func redactValue(v any, depth int) any {
	if depth >= maxRedactDepth {
		return "[TRUNCATED]"
	}
	switch val := v.(type) {
	case string:
		return redactString(val)
	case map[string]any:
		return redactMap(val, depth+1)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = redactValue(item, depth)
		}
		return out
	default:
		return v
	}
}

func redactString(s string) string {
	s = emailRe.ReplaceAllString(s, "[EMAIL_REDACTED]")
	s = ssnRe.ReplaceAllString(s, "[SSN_REDACTED]")
	s = phoneRe.ReplaceAllString(s, "[PHONE_REDACTED]")
	return s
}

// TruncateForLog shortens s to at most maxLen runes and strips ASCII control
// characters (< 0x20 or 0x7f), making the result safe to write to structured
// log fields. If maxLen is zero or negative, returns an empty string.
func TruncateForLog(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count >= maxLen {
			break
		}
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
		count++
	}
	return b.String()
}

// SanitizeHeader removes carriage returns (\r) and line feeds (\n) from s,
// preventing HTTP header injection (CRLF injection). The result is safe for
// use as an HTTP header value.
func SanitizeHeader(s string) string {
	return strings.NewReplacer("\r\n", "", "\r", "", "\n", "").Replace(s)
}

// IsRelativeURL returns true if rawURL is a relative path (starts with /)
// and does not point to a different host. This prevents open redirect
// vulnerabilities when using user-supplied redirect targets.
func IsRelativeURL(rawURL string) bool {
if rawURL == "" || rawURL[0] != '/' {
return false
}
// Reject protocol-relative URLs like //evil.com
if len(rawURL) > 1 && rawURL[1] == '/' {
return false
}
u, err := url.Parse(rawURL)
if err != nil {
return false
}
return u.Host == "" && u.Scheme == ""
}
