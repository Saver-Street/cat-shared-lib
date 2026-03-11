package middleware

import "net/http"

// SecureHeaders returns middleware that sets common security-related HTTP
// response headers. These headers protect against clickjacking, MIME-type
// sniffing, and cross-site scripting when combined with other measures.
//
// Headers set:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - Referrer-Policy: strict-origin-when-cross-origin
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
