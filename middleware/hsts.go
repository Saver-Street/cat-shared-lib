package middleware

import (
	"fmt"
	"net/http"
)

// HSTS returns middleware that sets the Strict-Transport-Security header,
// instructing browsers to only access the site via HTTPS for the given
// duration in seconds. Pass includeSubDomains=true to extend the policy
// to all subdomains.
func HSTS(maxAgeSeconds int, includeSubDomains bool) func(http.Handler) http.Handler {
	value := fmt.Sprintf("max-age=%d", maxAgeSeconds)
	if includeSubDomains {
		value += "; includeSubDomains"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", value)
			next.ServeHTTP(w, r)
		})
	}
}
