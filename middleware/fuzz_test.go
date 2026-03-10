package middleware

import (
	"net/http"
	"testing"
)

func FuzzGetClientIP(f *testing.F) {
	f.Add("203.0.113.5, 10.0.0.1", "192.168.1.1:8080")
	f.Add("", "10.0.0.5:8080")
	f.Add("  192.168.1.1  , 10.0.0.1", "127.0.0.1:80")
	f.Add("2001:db8::1, 10.0.0.1", "[::1]:8080")
	f.Add("", "192.168.1.1")

	f.Fuzz(func(t *testing.T, xff, remoteAddr string) {
		r, _ := http.NewRequest("GET", "/", nil)
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		r.RemoteAddr = remoteAddr

		ip := GetClientIP(r)
		// Should never return empty
		if ip == "" {
			t.Errorf("GetClientIP returned empty for xff=%q remote=%q", xff, remoteAddr)
		}
	})
}

func FuzzIsExemptFromRateLimit(f *testing.F) {
	f.Add("/assets/main.js")
	f.Add("/health")
	f.Add("/api/health")
	f.Add("/api/users")
	f.Add("/icons/logo.png")
	f.Add("/static/bundle.js")
	f.Add("")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic
		IsExemptFromRateLimit(path)
	})
}
