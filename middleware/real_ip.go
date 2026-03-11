package middleware

import (
	"net"
	"net/http"
	"strings"
)

// RealIP is middleware that sets the X-Real-IP header on the request to
// the client's real IP address, extracted from X-Forwarded-For or
// X-Real-IP headers, falling back to the remote address.
//
// This should be placed early in the middleware chain so that
// downstream handlers can rely on request.ClientIP or read
// X-Real-IP directly.
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if ip != "" {
			r.Header.Set("X-Real-IP", ip)
		}
		next.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the chain is the original client.
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
