package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// TokenBucketConfig configures the per-IP token bucket rate limiter.
type TokenBucketConfig struct {
	// Rate is the number of tokens added per second. Default: 10.
	Rate float64
	// Burst is the maximum number of tokens (bucket capacity). Default: 20.
	Burst int
	// CleanupInterval is how often stale entries are removed. Default: 5 minutes.
	CleanupInterval time.Duration
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

// TokenBucketLimiter is a per-IP token bucket rate limiter.
type TokenBucketLimiter struct {
	config  TokenBucketConfig
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	stopCh  chan struct{}
	stopped bool
}

// NewTokenBucketLimiter creates a new per-IP token bucket rate limiter.
func NewTokenBucketLimiter(cfg TokenBucketConfig) *TokenBucketLimiter {
	if cfg.Rate <= 0 {
		cfg.Rate = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}

	tbl := &TokenBucketLimiter{
		config:  cfg,
		buckets: make(map[string]*tokenBucket),
		stopCh:  make(chan struct{}),
	}

	go tbl.cleanupLoop()
	return tbl
}

// Allow checks whether a request from the given IP should be allowed.
// It consumes one token and returns true if the request is allowed.
func (tbl *TokenBucketLimiter) Allow(ip string) bool {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	now := time.Now()
	b, exists := tbl.buckets[ip]
	if !exists {
		b = &tokenBucket{
			tokens:   float64(tbl.config.Burst) - 1,
			lastTime: now,
		}
		tbl.buckets[ip] = b
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * tbl.config.Rate
	if b.tokens > float64(tbl.config.Burst) {
		b.tokens = float64(tbl.config.Burst)
	}
	b.lastTime = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Middleware returns an HTTP middleware that rate-limits requests using
// the token bucket algorithm. Exempt paths (health, static assets) bypass
// the limiter.
func (tbl *TokenBucketLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsExemptFromRateLimit(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		ip := GetClientIP(r)
		if !tbl.Allow(ip) {
			response.TooManyRequests(w, "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Stop terminates the background cleanup goroutine.
func (tbl *TokenBucketLimiter) Stop() {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()
	if !tbl.stopped {
		close(tbl.stopCh)
		tbl.stopped = true
	}
}

func (tbl *TokenBucketLimiter) cleanupLoop() {
	ticker := time.NewTicker(tbl.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-tbl.stopCh:
			return
		case <-ticker.C:
			tbl.cleanup()
		}
	}
}

func (tbl *TokenBucketLimiter) cleanup() {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	now := time.Now()
	for ip, b := range tbl.buckets {
		// Remove entries that have been idle long enough to fully refill
		idle := now.Sub(b.lastTime).Seconds()
		if idle > float64(tbl.config.Burst)/tbl.config.Rate*2 {
			delete(tbl.buckets, ip)
		}
	}
}
