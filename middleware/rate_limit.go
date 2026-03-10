package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// RateLimiterConfig configures the sliding-window rate limiter.
type RateLimiterConfig struct {
	// RequestsPerWindow is the maximum number of requests allowed per IP per window.
	// Defaults to 100 if zero or negative.
	RequestsPerWindow int
	// WindowDuration is the sliding time window over which requests are counted.
	// Defaults to 1 minute if zero or negative.
	WindowDuration time.Duration
	// CleanupInterval controls how often stale visitor entries are removed.
	// Defaults to WindowDuration * 2 if zero.
	CleanupInterval time.Duration
}

type visitor struct {
	count    int
	windowAt time.Time
}

// RateLimiter implements a per-IP sliding-window rate limiter.
type RateLimiter struct {
	config   RateLimiterConfig
	visitors map[string]*visitor
	mu       sync.RWMutex
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewRateLimiter creates a rate limiter with automatic stale-entry cleanup.
// Zero or negative RequestsPerWindow defaults to 100.
// Zero or negative WindowDuration defaults to 1 minute.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	if cfg.RequestsPerWindow <= 0 {
		cfg.RequestsPerWindow = 100
	}
	if cfg.WindowDuration <= 0 {
		cfg.WindowDuration = time.Minute
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = cfg.WindowDuration * 2
	}
	rl := &RateLimiter{
		config:   cfg,
		visitors: make(map[string]*visitor),
		stopCh:   make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Stop terminates the cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() { close(rl.stopCh) })
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now().UTC()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for ip, v := range rl.visitors {
		if now.Sub(v.windowAt) > rl.config.WindowDuration {
			delete(rl.visitors, ip)
		}
	}
}

// GetClientIP extracts the client IP from the request, preferring X-Forwarded-For.
// WARNING: X-Forwarded-For can be spoofed by clients. Only rely on this for
// best-effort rate limiting or logging. For security-critical access control,
// deploy behind a trusted reverse proxy that sets this header authoritatively.
func GetClientIP(r *http.Request) string {
	return clientIP(r)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// IsExemptFromRateLimit returns true for paths that must never be rate-limited.
func IsExemptFromRateLimit(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/icons/") ||
		strings.HasPrefix(path, "/static/") ||
		path == "/api/health" ||
		path == "/health"
}

// Middleware returns an http.Handler that enforces rate limits per IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsExemptFromRateLimit(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now().UTC()
		if !exists || now.Sub(v.windowAt) > rl.config.WindowDuration {
			rl.visitors[ip] = &visitor{count: 1, windowAt: now}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		if v.count > rl.config.RequestsPerWindow {
			remaining := rl.config.WindowDuration - now.Sub(v.windowAt)
			rl.mu.Unlock()
			slog.Warn("rate: limit exceeded", "ip", ip, "path", r.URL.Path, "count", v.count)
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", remaining.Seconds()+1))
			response.TooManyRequests(w, "Too many requests")
			return
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
