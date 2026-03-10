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

func TestGetClientIP_IPv6XFF(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "2001:db8::1, 10.0.0.1")
	if ip := GetClientIP(r); ip != "2001:db8::1" {
		t.Errorf("IPv6 X-Forwarded-For = %q, want 2001:db8::1", ip)
	}
}

func TestGetClientIP_IPv6RemoteAddr(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "[::1]:8080"
	if ip := GetClientIP(r); ip != "::1" {
		t.Errorf("IPv6 RemoteAddr = %q, want ::1", ip)
	}
}

func TestGetClientIP_XFFWithSpaces(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "  192.168.1.1  , 10.0.0.1")
	if ip := GetClientIP(r); ip != "192.168.1.1" {
		t.Errorf("XFF with spaces = %q, want 192.168.1.1", ip)
	}
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	rl := newTestRL(100)
	defer rl.Stop()

	done := make(chan int, 50)
	for i := 0; i < 50; i++ {
		go func() {
			code := doRequest(rl, "concurrent-ip")
			done <- code
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
	// No race, no panic — verified by -race flag
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

func TestGetClientIP_XFFEmptyFirstPart(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", " , 10.0.0.1")
	// Empty first part should fall through to RemoteAddr
	r.RemoteAddr = "192.168.1.1:8080"
	if ip := GetClientIP(r); ip != "192.168.1.1" {
		t.Errorf("XFF empty first part = %q, want 192.168.1.1", ip)
	}
}

func TestGetClientIP_NoPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5"
	if ip := GetClientIP(r); ip != "10.0.0.5" {
		t.Errorf("no-port RemoteAddr = %q, want 10.0.0.5", ip)
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	// First request passes
	if code := doRequest(rl, "reset-ip"); code != http.StatusOK {
		t.Errorf("first request: %d", code)
	}
	// Second request blocked
	if code := doRequest(rl, "reset-ip"); code != http.StatusTooManyRequests {
		t.Errorf("second request: %d, want 429", code)
	}
	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)
	// Third request should pass after window reset
	if code := doRequest(rl, "reset-ip"); code != http.StatusOK {
		t.Errorf("post-window request: %d, want 200", code)
	}
}

func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	doRequest(rl, "header-ip")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
	r.RemoteAddr = "header-ip:0"
	rl.Middleware(next).ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header missing")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: 5,
		WindowDuration:    50 * time.Millisecond,
		CleanupInterval:   50 * time.Millisecond,
	})
	defer rl.Stop()
	doRequest(rl, "cleanup-ip")
	// Wait for cleanup to remove stale entries
	time.Sleep(200 * time.Millisecond)
	rl.mu.RLock()
	_, exists := rl.visitors["cleanup-ip"]
	rl.mu.RUnlock()
	if exists {
		t.Error("stale visitor entry should have been cleaned up")
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
