package middleware

import "net/http"

// NoCache returns middleware that sets response headers to prevent caching.
// Useful for API endpoints where responses must not be stored by browsers
// or intermediate proxies.
//
// Headers set:
//   - Cache-Control: no-store, no-cache, must-revalidate
//   - Pragma: no-cache
//   - Expires: 0
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
