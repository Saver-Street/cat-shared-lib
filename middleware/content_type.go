package middleware

import (
	"net/http"
	"strings"
)

// ContentType returns middleware that enforces the Content-Type header on
// requests with a body (POST, PUT, PATCH). If the request's Content-Type
// does not match any of the allowed types, it responds with 415 Unsupported
// Media Type. GET, HEAD, DELETE, and OPTIONS requests are passed through.
func ContentType(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			ct := r.Header.Get("Content-Type")
			// Strip parameters (e.g. charset)
			if i := strings.IndexByte(ct, ';'); i >= 0 {
				ct = strings.TrimSpace(ct[:i])
			}

			for _, a := range allowed {
				if strings.EqualFold(ct, a) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		})
	}
}
