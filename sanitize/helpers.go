package sanitize

import (
	"net/url"
	"strings"
	"unicode"
)

// SanitizePhone strips all non-digit characters except a leading '+' from a
// phone number string.  The result contains only digits and an optional leading
// '+'.  An empty input returns an empty string.
func SanitizePhone(phone string) string {
	if phone == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(phone))
	for i, r := range phone {
		if r == '+' && i == 0 {
			b.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeURL trims whitespace from the URL, lowercases the scheme and host,
// and removes any trailing slash from the path (except for the root path "/").
// If the input is not a valid URL, the trimmed input is returned unchanged.
func SanitizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" {
		return rawURL
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	result := u.String()
	if len(result) > 1 && strings.HasSuffix(result, "/") && u.Path != "/" {
		result = strings.TrimRight(result, "/")
	}
	return result
}

// SanitizeName trims whitespace, collapses consecutive spaces into one, and
// title-cases each word.  Useful for cleaning up human names before storage.
func SanitizeName(name string) string {
	name = NormalizeWhitespace(name)
	words := strings.Fields(name)
	for i, w := range words {
		runes := []rune(w)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
