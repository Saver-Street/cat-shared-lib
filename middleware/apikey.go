package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// APIKey returns middleware that validates an API key sent in the specified
// header (e.g., "X-API-Key" or "Authorization"). The key is compared in
// constant time to prevent timing attacks. On failure the middleware replies
// with 401 Unauthorized.
func APIKey(header, key string) func(http.Handler) http.Handler {
	want := sha256.Sum256([]byte(key))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := sha256.Sum256([]byte(r.Header.Get(header)))
			if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyQuery returns middleware that validates an API key from a query
// parameter (e.g., "?api_key=secret"). The key is compared in constant
// time. On failure the middleware replies with 401 Unauthorized.
func APIKeyQuery(param, key string) func(http.Handler) http.Handler {
	want := sha256.Sum256([]byte(key))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := sha256.Sum256([]byte(r.URL.Query().Get(param)))
			if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyMulti returns middleware that accepts any of the provided keys sent
// in the specified header. Each key is compared in constant time. On failure
// the middleware replies with 401 Unauthorized.
func APIKeyMulti(header string, keys []string) func(http.Handler) http.Handler {
	hashes := make([][32]byte, len(keys))
	for i, k := range keys {
		hashes[i] = sha256.Sum256([]byte(k))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := sha256.Sum256([]byte(r.Header.Get(header)))
			valid := false
			for _, want := range hashes {
				if subtle.ConstantTimeCompare(got[:], want[:]) == 1 {
					valid = true
				}
			}
			if !valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
