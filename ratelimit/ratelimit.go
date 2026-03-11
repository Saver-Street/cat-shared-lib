// Package ratelimit provides a thread-safe, per-key token bucket rate limiter
// with automatic stale-entry cleanup. It is suitable for per-IP HTTP rate
// limiting as well as any use-case requiring per-key throttling.
package ratelimit

import (
	"sync"
	"time"
)

// Config configures a token bucket rate limiter.
type Config struct {
	// Rate is the number of tokens added per second. Default: 10.
	Rate float64
	// Burst is the maximum tokens (bucket capacity). Default: 20.
	Burst int
	// CleanupInterval controls how often idle buckets are removed.
	// Default: 5 minutes.
	CleanupInterval time.Duration
	// MaxIdleTime is how long a bucket can be idle before removal.
	// Default: 10 minutes.
	MaxIdleTime time.Duration
}

func (c *Config) defaults() {
	if c.Rate <= 0 {
		c.Rate = 10
	}
	if c.Burst <= 0 {
		c.Burst = 20
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = 5 * time.Minute
	}
	if c.MaxIdleTime <= 0 {
		c.MaxIdleTime = 10 * time.Minute
	}
}

type bucket struct {
	tokens   float64
	lastTime time.Time
}

// Limiter is a per-key token bucket rate limiter.
type Limiter struct {
	config  Config
	mu      sync.Mutex
	buckets map[string]*bucket
	stopCh  chan struct{}
	stopped bool
	now     func() time.Time // injectable clock for testing
}

// New creates a Limiter with the given configuration and starts a background
// goroutine for stale-entry cleanup. Call Stop when the limiter is no longer needed.
func New(cfg Config) *Limiter {
	cfg.defaults()
	l := &Limiter{
		config:  cfg,
		buckets: make(map[string]*bucket),
		stopCh:  make(chan struct{}),
		now:     time.Now,
	}
	go l.cleanupLoop()
	return l
}

// Allow checks whether a request for the given key should be allowed.
// It consumes one token and returns true if the request is permitted.
func (l *Limiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

// AllowN checks whether n tokens are available for the given key.
// It consumes n tokens if available and returns true, or returns false
// without consuming any tokens.
func (l *Limiter) AllowN(key string, n int) bool {
	if n <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, exists := l.buckets[key]
	if !exists {
		initial := float64(l.config.Burst) - float64(n)
		if initial < 0 {
			return false
		}
		l.buckets[key] = &bucket{
			tokens:   initial,
			lastTime: now,
		}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * l.config.Rate
	if b.tokens > float64(l.config.Burst) {
		b.tokens = float64(l.config.Burst)
	}
	b.lastTime = now

	if b.tokens < float64(n) {
		return false
	}
	b.tokens -= float64(n)
	return true
}

// Stop halts the background cleanup goroutine. Safe to call multiple times.
func (l *Limiter) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.stopped {
		l.stopped = true
		close(l.stopCh)
	}
}

// Len returns the current number of tracked keys.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	for key, b := range l.buckets {
		if now.Sub(b.lastTime) > l.config.MaxIdleTime {
			delete(l.buckets, key)
		}
	}
}

// Remaining returns the current number of available tokens for the given key
// without consuming any. If the key has no bucket yet, it returns the burst
// capacity.
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, exists := l.buckets[key]
	if !exists {
		return l.config.Burst
	}

	now := l.now()
	elapsed := now.Sub(b.lastTime).Seconds()
	tokens := b.tokens + elapsed*l.config.Rate
	if tokens > float64(l.config.Burst) {
		tokens = float64(l.config.Burst)
	}
	return int(tokens)
}

// Reset removes the bucket for the given key, restoring it to full capacity
// on next access.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}
