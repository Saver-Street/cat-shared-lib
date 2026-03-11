package middleware

import "net/http"

// MaxBody returns middleware that limits the size of incoming request bodies.
// If the body exceeds maxBytes, the server returns 413 Request Entity Too Large.
// This protects against denial-of-service attacks that send oversized payloads.
func MaxBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
