package middleware

import (
	"testing"
	"time"
)

func newTestBF(max int) *BruteForceGuard {
	return NewBruteForceGuard(BruteForceConfig{
		MaxFailures:   max,
		BlockDuration: 100 * time.Millisecond,
		Window:        200 * time.Millisecond,
	})
}

func TestBruteForce_NotBlockedInitially(t *testing.T) {
	g := newTestBF(3)
	defer g.Stop()
	if g.IsBlocked("1.1.1.1") {
		t.Error("IP should not be blocked initially")
	}
}

func TestBruteForce_BlocksAfterMaxFailures(t *testing.T) {
	g := newTestBF(3)
	defer g.Stop()
	ip := "2.2.2.2"
	g.RecordFailure(ip)
	g.RecordFailure(ip)
	blocked := g.RecordFailure(ip)
	if !blocked {
		t.Error("3rd failure should have triggered block")
	}
	if !g.IsBlocked(ip) {
		t.Error("IP should be blocked after max failures")
	}
}

func TestBruteForce_ResetClearsBlock(t *testing.T) {
	g := newTestBF(2)
	defer g.Stop()
	ip := "3.3.3.3"
	g.RecordFailure(ip)
	g.RecordFailure(ip)
	g.Reset(ip)
	if g.IsBlocked(ip) {
		t.Error("IP should not be blocked after Reset")
	}
}

func TestBruteForce_UnblocksAfterDuration(t *testing.T) {
	g := newTestBF(1)
	defer g.Stop()
	ip := "4.4.4.4"
	g.RecordFailure(ip)
	if !g.IsBlocked(ip) {
		t.Fatal("should be blocked after 1 failure (maxFailures=1)")
	}
	time.Sleep(150 * time.Millisecond)
	if g.IsBlocked(ip) {
		t.Error("should be unblocked after blockDuration expires")
	}
}

func TestBruteForce_IndependentIPs(t *testing.T) {
	g := newTestBF(2)
	defer g.Stop()
	g.RecordFailure("5.5.5.5")
	g.RecordFailure("5.5.5.5")
	if !g.IsBlocked("5.5.5.5") {
		t.Error("5.5.5.5 should be blocked")
	}
	if g.IsBlocked("6.6.6.6") {
		t.Error("6.6.6.6 should not be blocked")
	}
}
