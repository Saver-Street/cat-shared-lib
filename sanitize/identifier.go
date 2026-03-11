package sanitize

import (
	"strings"
	"unicode"
)

// SQLIdentifier sanitises s so it is safe to use as a SQL identifier
// (table or column name). It removes all characters except letters,
// digits, and underscores, and trims leading digits. An empty result
// returns "_".
func SQLIdentifier(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || r == '_' || (b.Len() > 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}

// CSVEscape escapes a string for safe inclusion in a CSV field. If the
// value contains a comma, double-quote, or newline, it is quoted and
// internal double-quotes are doubled.
func CSVEscape(s string) string {
	needsQuoting := strings.ContainsAny(s, ",\"\r\n")
	if !needsQuoting {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		if r == '"' {
			b.WriteString(`""`)
		} else {
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// HeaderName normalises an HTTP header name to canonical form
// (e.g. "content-type" → "Content-Type"). It strips non-printable
// characters and collapses hyphens.
func HeaderName(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	upper := true
	prev := false // previous was hyphen
	for _, r := range s {
		if r == '-' {
			if prev || b.Len() == 0 {
				continue
			}
			b.WriteByte('-')
			upper = true
			prev = true
			continue
		}
		prev = false
		if !unicode.IsPrint(r) {
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	// Trim trailing hyphen
	result := b.String()
	return strings.TrimRight(result, "-")
}

// EnvVarName converts s into a valid environment variable name by
// upper-casing letters, replacing non-alphanumeric characters with
// underscores, collapsing runs, and trimming leading digits.
func EnvVarName(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevUnderscore := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
			prevUnderscore = false
		} else if unicode.IsDigit(r) && b.Len() > 0 {
			b.WriteRune(r)
			prevUnderscore = false
		} else if b.Len() > 0 && !prevUnderscore {
			b.WriteByte('_')
			prevUnderscore = true
		}
	}
	result := b.String()
	return strings.TrimRight(result, "_")
}
