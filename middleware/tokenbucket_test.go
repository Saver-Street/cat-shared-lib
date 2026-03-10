package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucket_DefaultConfig(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{})
	defer tbl.Stop()

	if tbl.config.Rate != 10 {
		t.Errorf("default rate = %f, want 10", tbl.config.Rate)
	}
	if tbl.config.Burst != 20 {
		t.Errorf("default burst = %d, want 20", tbl.config.Burst)
	}
	if tbl.config.CleanupInterval != 5*time.Minute {
		t.Errorf("default cleanup = %v, want 5m", tbl.config.CleanupInterval)
	}
}

func TestTokenBucket_AllowWithinBurst(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 10, Burst: 5})
	defer tbl.Stop()

	for i := range 5 {
		if !tbl.Allow("1.2.3.4") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestTokenBucket_DenyOverBurst(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 3})
	defer tbl.Stop()

	for range 3 {
		tbl.Allow("1.2.3.4")
	}

	if tbl.Allow("1.2.3.4") {
		t.Error("4th request should be denied after burst of 3")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 100, Burst: 2})
	defer tbl.Stop()

	tbl.Allow("1.2.3.4")
	tbl.Allow("1.2.3.4")

	if tbl.Allow("1.2.3.4") {
		t.Error("should be denied before refill")
	}

	time.Sleep(50 * time.Millisecond) // 100 tokens/s * 0.05s = 5 tokens

	if !tbl.Allow("1.2.3.4") {
		t.Error("should be allowed after refill")
	}
}

func TestTokenBucket_PerIP(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{Rate: 1, Burst: 1})
	defer tbl.Stop()

	if !tbl.Allow("1.1.1.1") {
		t.Error("first IP first request should be allowed")
	}
	if !tbl.Allow("2.2.2.2") {
		t.Error("second IP first request should be allowed")
	}
	if tbl.Allow("1.1.1.1") {
		t.Error("first IP second request should be denied")
	}
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

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rr.Code)
	}
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
	if rr.Code != http.StatusOK {
		t.Fatalf("first request: got %d, want 200", rr.Code)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request: got %d, want 429", rr.Code)
	}
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
	if rr.Code != http.StatusOK {
		t.Errorf("exempt path got %d, want 200", rr.Code)
	}
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

	if exists {
		t.Error("stale entry should be cleaned up")
	}
}

func TestTokenBucket_StopIdempotent(t *testing.T) {
	tbl := NewTokenBucketLimiter(TokenBucketConfig{})
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
