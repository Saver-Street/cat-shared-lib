package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestTokenBucket_DefaultConfig(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{})
	defer tbl.Stop()

	testkit.AssertEqual(t, tbl.config.Rate, float64(10))
	testkit.AssertEqual(t, tbl.config.Burst, 20)
	testkit.AssertEqual(t, tbl.config.CleanupInterval, 5*time.Minute)
}

func TestTokenBucket_AllowWithinBurst(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 10, Burst: 5})
	defer tbl.Stop()

	for range 5 {
		testkit.AssertTrue(t, tbl.Allow("1.2.3.4"))
	}
}

func TestTokenBucket_DenyOverBurst(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 3})
	defer tbl.Stop()

	for range 3 {
		tbl.Allow("1.2.3.4")
	}

	testkit.AssertFalse(t, tbl.Allow("1.2.3.4"))
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 100, Burst: 2})
	defer tbl.Stop()

	tbl.Allow("1.2.3.4")
	tbl.Allow("1.2.3.4")

	testkit.AssertFalse(t, tbl.Allow("1.2.3.4"))

	time.Sleep(50 * time.Millisecond) // 100 tokens/s * 0.05s = 5 tokens

	testkit.AssertTrue(t, tbl.Allow("1.2.3.4"))
}

func TestTokenBucket_PerIP(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 1})
	defer tbl.Stop()

	testkit.AssertTrue(t, tbl.Allow("1.1.1.1"))
	testkit.AssertTrue(t, tbl.Allow("2.2.2.2"))
	testkit.AssertFalse(t, tbl.Allow("1.1.1.1"))
}

func TestTokenBucket_Middleware_OK(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 10, Burst: 10})
	defer tbl.Stop()

	handler := tbl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "5.5.5.5:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestTokenBucket_Middleware_RateLimited(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 1})
	defer tbl.Stop()

	handler := tbl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "6.6.6.6:1234"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.RequireEqual(t, rr.Code, http.StatusOK)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusTooManyRequests)
}

func TestTokenBucket_Middleware_ExemptPath(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 1})
	defer tbl.Stop()

	handler := tbl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use burst
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "7.7.7.7:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Exempt path should bypass
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "7.7.7.7:1234"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestTokenBucket_Cleanup(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1000, Burst: 1})
	defer tbl.Stop()

	tbl.Allow("old-ip")

	// Manually set lastTime to far past
	tbl.mu.Lock()
	tbl.buckets["old-ip"].lastTime = time.Now().Add(-time.Hour)
	tbl.mu.Unlock()

	tbl.cleanup()

	tbl.mu.Lock()
	_, exists := tbl.buckets["old-ip"]
	tbl.mu.Unlock()

	testkit.AssertFalse(t, exists)
}

func TestTokenBucket_StopIdempotent(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{CleanupInterval: 10 * time.Millisecond})
	// Let the cleanup ticker fire at least once.
	time.Sleep(50 * time.Millisecond)
	tbl.Stop()
	tbl.Stop() // should not panic
}

func TestTokenBucket_BurstCap(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1000, Burst: 3})
	defer tbl.Stop()

	tbl.Allow("cap-ip")
	// Wait for many tokens to accrue
	time.Sleep(50 * time.Millisecond) // would accrue 50 tokens at 1000/s

	// But burst cap is 3, so only 3 allowed
	count := 0
	for range 10 {
		if tbl.Allow("cap-ip") {
			count++
		}
	}
	if count > 3 {
		t.Errorf("burst cap: %d requests allowed, want <= 3", count)
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1000000, Burst: 1000000})
	defer tbl.Stop()
	for b.Loop() {
		tbl.Allow("bench-ip")
	}
}
