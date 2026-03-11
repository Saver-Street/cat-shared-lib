package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
		testkit.AssertEqual(t, doRequest(rl, "1.2.3.4"), http.StatusOK)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newTestRL(2)
	defer rl.Stop()
	doRequest(rl, "5.6.7.8")
	doRequest(rl, "5.6.7.8")
	testkit.AssertEqual(t, doRequest(rl, "5.6.7.8"), http.StatusTooManyRequests)
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
		testkit.AssertStatus(t, w, http.StatusOK)
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	testkit.AssertEqual(t, doRequest(rl, "10.0.0.1"), http.StatusOK)
	testkit.AssertEqual(t, doRequest(rl, "10.0.0.2"), http.StatusOK)
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	testkit.AssertEqual(t, GetClientIP(r), "203.0.113.5")
}

func TestGetClientIP_IPv6XFF(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "2001:db8::1, 10.0.0.1")
	testkit.AssertEqual(t, GetClientIP(r), "2001:db8::1")
}

func TestGetClientIP_IPv6RemoteAddr(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "[::1]:8080"
	testkit.AssertEqual(t, GetClientIP(r), "::1")
}

func TestGetClientIP_XFFWithSpaces(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "  192.168.1.1  , 10.0.0.1")
	testkit.AssertEqual(t, GetClientIP(r), "192.168.1.1")
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
		testkit.AssertTrue(t, IsExemptFromRateLimit(p))
	}
	notExempt := []string{"/api/apply", "/api/user", "/"}
	for _, p := range notExempt {
		testkit.AssertFalse(t, IsExemptFromRateLimit(p))
	}
}

// --- Benchmarks ---

func TestGetClientIP_XFFEmptyFirstPart(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", " , 10.0.0.1")
	// Empty first part should fall through to RemoteAddr
	r.RemoteAddr = "192.168.1.1:8080"
	testkit.AssertEqual(t, GetClientIP(r), "192.168.1.1")
}

func TestGetClientIP_NoPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5"
	testkit.AssertEqual(t, GetClientIP(r), "10.0.0.5")
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	// First request passes
	testkit.AssertEqual(t, doRequest(rl, "reset-ip"), http.StatusOK)
	// Second request blocked
	testkit.AssertEqual(t, doRequest(rl, "reset-ip"), http.StatusTooManyRequests)
	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)
	// Third request should pass after window reset
	testkit.AssertEqual(t, doRequest(rl, "reset-ip"), http.StatusOK)
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
	testkit.RequireEqual(t, w.Code, http.StatusTooManyRequests)
	testkit.AssertNotEqual(t, w.Header().Get("Retry-After"), "")
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
	testkit.AssertFalse(t, exists)
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

func TestNewRateLimiter_ZeroConfigDefaults(t *testing.T) {
	// Zero RequestsPerWindow and WindowDuration should apply sane defaults.
	rl := NewRateLimiter(RateLimiterConfig{})
	defer rl.Stop()

	testkit.AssertEqual(t, rl.config.RequestsPerWindow, 100)
	testkit.AssertEqual(t, rl.config.WindowDuration, time.Minute)
	testkit.AssertEqual(t, rl.config.CleanupInterval, 2*time.Minute)
}

func TestNewRateLimiter_NegativeConfigDefaults(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{RequestsPerWindow: -5, WindowDuration: -1})
	defer rl.Stop()

	testkit.AssertEqual(t, rl.config.RequestsPerWindow, 100)
	testkit.AssertEqual(t, rl.config.WindowDuration, time.Minute)
}
