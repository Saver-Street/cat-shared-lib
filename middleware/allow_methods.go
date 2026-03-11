package middleware

import (
	"net/http"
	"strings"
)

// AllowMethods returns middleware that rejects requests whose HTTP method is
// not in the allowed set, responding with 405 Method Not Allowed. The Allow
// header is set to the comma-separated list of allowed methods.
func AllowMethods(methods ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(methods))
	for _, m := range methods {
		allowed[strings.ToUpper(m)] = true
	}
	allow := strings.Join(methods, ", ")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowed[r.Method] {
				w.Header().Set("Allow", allow)
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
