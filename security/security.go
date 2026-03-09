// Package security provides input validation and PII redaction helpers.
package security

import (
	"regexp"
	"strings"
)

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

var piiFields = map[string]bool{
	"email": true, "phone": true, "address": true, "ssn": true,
	"password": true, "resume": true, "socialSecurityNumber": true,
	"phoneNumber": true, "emailAddress": true, "streetAddress": true,
	"zipCode": true, "postalCode": true, "dateOfBirth": true,
}

var (
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phoneRe = regexp.MustCompile(`(\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	ssnRe   = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)
)

// RedactPII performs server-side PII scrubbing on audit payloads.
func RedactPII(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))
	for k, v := range data {
		if piiFields[k] || piiFields[strings.ToLower(k)] {
			result[k] = "[REDACTED]"
			continue
		}
		result[k] = redactValue(v)
	}
	return result
}

func redactValue(v any) any {
	switch val := v.(type) {
	case string:
		return redactString(val)
	case map[string]any:
		return RedactPII(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = redactValue(item)
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
