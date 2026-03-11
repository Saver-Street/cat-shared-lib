package urlutil

import (
	"net/url"
	"path"
	"strings"
)

// Join joins a base URL with path segments, properly handling slashes.
func Join(base string, segments ...string) string {
	if len(segments) == 0 {
		return base
	}
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	parts := append([]string{u.Path}, segments...)
	u.Path = path.Join(parts...)
	return u.String()
}

// SetQuery returns a copy of rawURL with the given query parameter set.
func SetQuery(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

// AddQuery returns a copy of rawURL with the given query parameter added.
func AddQuery(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Add(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

// RemoveQuery returns a copy of rawURL with the given query parameter removed.
func RemoveQuery(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()
	return u.String()
}

// StripQuery returns rawURL without any query string or fragment.
func StripQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// Domain extracts the domain (host without port) from a URL.
func Domain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	return host
}

// IsAbsolute returns true if the URL has a scheme (e.g., http://, https://).
func IsAbsolute(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// IsHTTPS returns true if the URL uses the https scheme.
func IsHTTPS(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Scheme, "https")
}

// HasQuery returns true if the URL contains the given query parameter.
func HasQuery(rawURL, key string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Query().Has(key)
}

// QueryValue returns the value of the given query parameter.
func QueryValue(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}
