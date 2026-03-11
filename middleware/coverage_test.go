package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// --- Additional brute force tests ---

func TestBruteForce_DefaultConfig(t *testing.T) {
	g := NewBruteForceGuard(BruteForceConfig{})
	defer g.Stop()
	testkit.AssertEqual(t, g.cfg.MaxFailures, 5)
	testkit.AssertEqual(t, g.cfg.BlockDuration, 15*time.Minute)
	testkit.AssertEqual(t, g.cfg.Window, 10*time.Minute)
}

func TestBruteForce_Middleware_BlockedIP(t *testing.T) {
	g := newTestBF(1)
	defer g.Stop()
	ip := "7.7.7.7"
	g.RecordFailure(ip)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("blocked IP should not reach next handler")
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/login", nil)
	r.RemoteAddr = ip + ":1234"
	g.Middleware(next).ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusTooManyRequests)
	testkit.AssertNotEqual(t, w.Header().Get("Retry-After"), "")
}

func TestBruteForce_Middleware_AllowedIP(t *testing.T) {
	g := newTestBF(5)
	defer g.Stop()
	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/login", nil)
	r.RemoteAddr = "8.8.8.8:9999"
	g.Middleware(next).ServeHTTP(w, r)
	testkit.AssertTrue(t, reached)
	testkit.AssertStatus(t, w, http.StatusOK)
}

func TestBruteForce_Cleanup_RemovesExpiredBlock(t *testing.T) {
	g := newTestBF(1)
	defer g.Stop()
	ip := "9.0.0.1"
	g.RecordFailure(ip)
	if !g.IsBlocked(ip) {
		t.Fatal("should be blocked")
	}
	time.Sleep(150 * time.Millisecond)
	g.cleanup()
	g.mu.Lock()
	_, exists := g.entries[ip]
	g.mu.Unlock()
	testkit.AssertFalse(t, exists)
}

func TestBruteForce_Cleanup_RemovesExpiredWindow(t *testing.T) {
	g := newTestBF(5)
	defer g.Stop()
	ip := "9.0.0.2"
	g.RecordFailure(ip) // 1 failure, not blocked
	if g.IsBlocked(ip) {
		t.Fatal("should not be blocked yet")
	}
	time.Sleep(250 * time.Millisecond) // wait > window (200ms)
	g.cleanup()
	g.mu.Lock()
	_, exists := g.entries[ip]
	g.mu.Unlock()
	testkit.AssertFalse(t, exists)
}

func TestBruteForce_RecordFailure_WindowExpiredResets(t *testing.T) {
	g := newTestBF(3)
	defer g.Stop()
	ip := "9.0.0.3"
	g.RecordFailure(ip)
	g.RecordFailure(ip)
	time.Sleep(250 * time.Millisecond) // window expires
	// After window expires, next RecordFailure should start fresh
	blocked := g.RecordFailure(ip)
	testkit.AssertFalse(t, blocked)
}

func TestBruteForce_IsBlocked_BlockExpires(t *testing.T) {
	// Manually create guard to control cleanup timing precisely
	g := &BruteForceGuard{
		cfg: BruteForceConfig{
			MaxFailures:   1,
			BlockDuration: 50 * time.Millisecond,
			Window:        200 * time.Millisecond,
		},
		entries: make(map[string]*bruteEntry),
		stopCh:  make(chan struct{}),
	}
	// Don't start cleanup goroutine — we want IsBlocked to detect expiry
	defer close(g.stopCh)

	ip := "9.0.0.4"
	g.RecordFailure(ip)
	if !g.IsBlocked(ip) {
		t.Fatal("should be blocked")
	}
	time.Sleep(80 * time.Millisecond) // > blockDuration
	testkit.AssertFalse(t, g.IsBlocked(ip))
	g.mu.Lock()
	_, exists := g.entries[ip]
	g.mu.Unlock()
	testkit.AssertFalse(t, exists)
}

// --- Additional JWT/context tests ---

func TestGetUserEmail_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetUserEmail(r), "")
}

func TestGetUserEmail_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), UserEmailKey, "test@example.com")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetUserEmail(r), "test@example.com")
}

func TestSetUserEmail_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserEmail(r.Context(), "a@b.com"))
	testkit.AssertEqual(t, GetUserEmail(r), "a@b.com")
}

func TestGetExtCandidateID_Empty(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	testkit.AssertEqual(t, GetExtCandidateID(r), "")
}

func TestGetExtCandidateID_Set(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(r.Context(), ExtCandidateIDKey, "ext-cand-xyz")
	r = r.WithContext(ctx)
	testkit.AssertEqual(t, GetExtCandidateID(r), "ext-cand-xyz")
}

func TestRequireAdmin_NoUserID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})
	handler := RequireAdmin(next)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	handler.ServeHTTP(w, r)
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestSetUserRole_RoundTrip(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(SetUserRole(r.Context(), "moderator"))
	testkit.AssertEqual(t, GetUserRole(r), "moderator")
}

// --- Additional rate limiter tests ---

func TestRateLimiter_Cleanup_RemovesStaleEntries(t *testing.T) {
	rl := newTestRL(10)
	defer rl.Stop()
	doRequest(rl, "11.0.0.1")
	time.Sleep(150 * time.Millisecond) // window expires (100ms)
	rl.cleanup()
	rl.mu.RLock()
	_, exists := rl.visitors["11.0.0.1"]
	rl.mu.RUnlock()
	testkit.AssertFalse(t, exists)
}

func TestRateLimiter_WindowResetsAfterDuration(t *testing.T) {
	rl := newTestRL(1)
	defer rl.Stop()
	if code := doRequest(rl, "12.0.0.1"); code != http.StatusOK {
		t.Fatalf("first request: %d", code)
	}
	if code := doRequest(rl, "12.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("second request (over limit): %d", code)
	}
	time.Sleep(150 * time.Millisecond) // wait for window to expire
	testkit.AssertEqual(t, doRequest(rl, "12.0.0.1"), http.StatusOK)
}

func TestClientIP_NoPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.168.1.1" // no port
	testkit.AssertEqual(t, GetClientIP(r), "192.168.1.1")
}

func TestClientIP_EmptyXFF(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "")
	r.RemoteAddr = "10.0.0.5:8080"
	testkit.AssertEqual(t, GetClientIP(r), "10.0.0.5")
}

func TestRateLimiter_DefaultCleanupInterval(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: 10,
		WindowDuration:    5 * time.Second,
	})
	defer rl.Stop()
	testkit.AssertEqual(t, rl.config.CleanupInterval, 10*time.Second)
}

func TestRateLimiter_ExemptPaths_Icons(t *testing.T) {
	testkit.AssertTrue(t, IsExemptFromRateLimit("/icons/favicon.ico"))
}

func TestRateLimiter_ExemptPaths_Static(t *testing.T) {
	testkit.AssertTrue(t, IsExemptFromRateLimit("/static/bundle.js"))
}

func TestBruteForce_IsBlocked_EntryExistsButNotBlocked(t *testing.T) {
	g := newTestBF(5)
	defer g.Stop()
	ip := "9.0.0.5"
	g.RecordFailure(ip) // 1 of 5, not blocked
	// Entry exists with blockedAt == nil
	testkit.AssertFalse(t, g.IsBlocked(ip))
}

func TestBruteForce_ConcurrentRecordFailure(t *testing.T) {
	g := newTestBF(100)
	defer g.Stop()
	ip := "10.0.0.1"

	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			g.RecordFailure(ip)
			done <- true
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// Should not panic or race — verified by -race flag
	testkit.AssertFalse(t, g.IsBlocked(ip))
}

func TestBruteForce_ConcurrentIsBlocked(t *testing.T) {
	g := newTestBF(1)
	defer g.Stop()
	ip := "10.0.0.2"
	g.RecordFailure(ip) // blocks immediately

	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			g.IsBlocked(ip)
			done <- true
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}

func BenchmarkBruteForce_RecordFailure(b *testing.B) {
	g := NewBruteForceGuard(BruteForceConfig{
		MaxFailures:   1000000,
		BlockDuration: time.Hour,
		Window:        time.Hour,
	})
	defer g.Stop()
	for b.Loop() {
		g.RecordFailure("bench-ip")
	}
}

func BenchmarkBruteForce_IsBlocked(b *testing.B) {
	g := NewBruteForceGuard(BruteForceConfig{
		MaxFailures:   1000000,
		BlockDuration: time.Hour,
		Window:        time.Hour,
	})
	defer g.Stop()
	g.RecordFailure("bench-ip")
	for b.Loop() {
		g.IsBlocked("bench-ip")
	}
}

func TestRateLimiter_CleanupLoop_FiresOnTicker(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: 100,
		WindowDuration:    50 * time.Millisecond,
		CleanupInterval:   80 * time.Millisecond,
	})
	defer rl.Stop()
	doRequest(rl, "13.0.0.1")
	// Wait for window to expire AND cleanup tick to fire
	time.Sleep(200 * time.Millisecond)
	rl.mu.RLock()
	_, exists := rl.visitors["13.0.0.1"]
	rl.mu.RUnlock()
	testkit.AssertFalse(t, exists)
}

func TestBruteForceGuard_StopIdempotent(t *testing.T) {
	g := newTestBF(5)
	// Calling Stop twice must not panic.
	g.Stop()
	g.Stop()
}

func TestRateLimiter_StopIdempotent(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerWindow: 10,
		WindowDuration:    time.Second,
	})
	// Calling Stop twice must not panic.
	rl.Stop()
	rl.Stop()
}
