package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestRL(n int) *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: n,
		WindowDuration:    100 * time.Millisecond,
		CleanupInterval:   200 * time.Millisecond,
	})
}

func doRequest(rl *RateLimiter, ip string) int {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
	r.RemoteAddr = ip + ":0"
	rl.Middleware(next).ServeHTTP(w, r)
	return w.Code
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := newTestRL(3)
	defer rl.Stop()
	for i := 0; i < 3; i++ {
		if code := doRequest(rl, "1.2.3.4"); code != http.StatusOK {
			t.Errorf("request %d: status = %d, want 200", i+1, code)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newTestRL(2)
	defer rl.Stop()
	doRequest(rl, "5.6.7.8")
	doRequest(rl, "5.6.7.8")
	if code := doRequest(rl, "5.6.7.8"); code != http.StatusTooManyRequests {
		t.Errorf("3rd request: status = %d, want 429", code)
	}
}

func TestRateLimiter_ExemptPaths(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	for _, path := range []string{"/assets/main.js", "/health", "/api/health"} {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, path, nil)
		r.RemoteAddr = "9.9.9.9:0"
		rl.Middleware(next).ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("exempt path %q: status = %d, want 200", path, w.Code)
		}
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	if code := doRequest(rl, "10.0.0.1"); code != http.StatusOK {
		t.Errorf("ip1 first: %d", code)
	}
	if code := doRequest(rl, "10.0.0.2"); code != http.StatusOK {
		t.Errorf("ip2 first: %d", code)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	if ip := GetClientIP(r); ip != "203.0.113.5" {
		t.Errorf("X-Forwarded-For IP = %q, want 203.0.113.5", ip)
	}
}

func TestIsExemptFromRateLimit(t *testing.T) {
	exempt := []string{"/assets/app.js", "/icons/logo.png", "/static/x", "/health", "/api/health"}
	for _, p := range exempt {
		if !IsExemptFromRateLimit(p) {
			t.Errorf("path %q should be exempt", p)
		}
	}
	notExempt := []string{"/api/apply", "/api/user", "/"}
	for _, p := range notExempt {
		if IsExemptFromRateLimit(p) {
			t.Errorf("path %q should NOT be exempt", p)
		}
	}
}

// --- Benchmarks ---

func BenchmarkGetClientIP(b *testing.B) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	for b.Loop() {
		GetClientIP(r)
	}
}

func BenchmarkIsExemptFromRateLimit(b *testing.B) {
	for b.Loop() {
		IsExemptFromRateLimit("/api/apply")
		IsExemptFromRateLimit("/assets/main.js")
		IsExemptFromRateLimit("/health")
	}
}
