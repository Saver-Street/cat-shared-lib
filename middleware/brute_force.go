package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// BruteForceConfig configures the brute-force login protection.
type BruteForceConfig struct {
	// MaxFailures is the number of consecutive failures before an IP is blocked.
	// Defaults to 5 if zero.
	MaxFailures int
	// BlockDuration is how long a blocked IP remains locked out.
	// Defaults to 15 minutes if zero.
	BlockDuration time.Duration
	// Window is the sliding time window in which failures are counted.
	// Defaults to 10 minutes if zero.
	Window time.Duration
}

type bruteEntry struct {
	failures  int
	firstSeen time.Time
	blockedAt *time.Time
}

// BruteForceGuard tracks failed login attempts per IP and blocks repeat offenders.
type BruteForceGuard struct {
	cfg      BruteForceConfig
	entries  map[string]*bruteEntry
	mu       sync.Mutex
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewBruteForceGuard creates a guard with automatic stale-entry cleanup.
func NewBruteForceGuard(cfg BruteForceConfig) *BruteForceGuard {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 5
	}
	if cfg.BlockDuration == 0 {
		cfg.BlockDuration = 15 * time.Minute
	}
	if cfg.Window == 0 {
		cfg.Window = 10 * time.Minute
	}
	g := &BruteForceGuard{
		cfg:     cfg,
		entries: make(map[string]*bruteEntry),
		stopCh:  make(chan struct{}),
	}
	go g.cleanupLoop()
	return g
}

// Stop terminates the cleanup goroutine. Safe to call multiple times.
func (g *BruteForceGuard) Stop() {
	g.stopOnce.Do(func() { close(g.stopCh) })
}

func (g *BruteForceGuard) cleanupLoop() {
	ticker := time.NewTicker(g.cfg.BlockDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			g.cleanup()
		case <-g.stopCh:
			return
		}
	}
}

func (g *BruteForceGuard) cleanup() {
	now := time.Now().UTC()
	g.mu.Lock()
	defer g.mu.Unlock()
	for ip, e := range g.entries {
		if e.blockedAt != nil && now.Sub(*e.blockedAt) > g.cfg.BlockDuration {
			delete(g.entries, ip)
			continue
		}
		if e.blockedAt == nil && now.Sub(e.firstSeen) > g.cfg.Window {
			delete(g.entries, ip)
		}
	}
}

// IsBlocked returns true if the IP is currently blocked.
func (g *BruteForceGuard) IsBlocked(ip string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	e, ok := g.entries[ip]
	if !ok || e.blockedAt == nil {
		return false
	}
	if time.Since(*e.blockedAt) > g.cfg.BlockDuration {
		delete(g.entries, ip)
		return false
	}
	return true
}

// RecordFailure records a failed attempt. Returns true if the IP was just blocked.
func (g *BruteForceGuard) RecordFailure(ip string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now().UTC()
	e, ok := g.entries[ip]
	if !ok || (e.blockedAt == nil && now.Sub(e.firstSeen) > g.cfg.Window) {
		newEntry := &bruteEntry{failures: 1, firstSeen: now}
		g.entries[ip] = newEntry
		if newEntry.failures >= g.cfg.MaxFailures {
			newEntry.blockedAt = &now
			slog.Warn("brute: IP blocked", "ip", ip, "failures", newEntry.failures)
			return true
		}
		return false
	}
	e.failures++
	if e.failures >= g.cfg.MaxFailures && e.blockedAt == nil {
		e.blockedAt = &now
		slog.Warn("brute: IP blocked", "ip", ip, "failures", e.failures)
		return true
	}
	return false
}

// Reset clears the failure count for an IP (call on successful login).
func (g *BruteForceGuard) Reset(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.entries, ip)
}

// Middleware blocks requests from IPs that have exceeded the failure threshold.
func (g *BruteForceGuard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if g.IsBlocked(ip) {
			slog.Warn("brute: request blocked", "ip", ip, "path", r.URL.Path)
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", g.cfg.BlockDuration.Seconds()))
			http.Error(w, `{"error":"Too many failed attempts. Try again later."}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
