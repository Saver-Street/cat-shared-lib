package middleware

import (
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
	testkit.AssertFalse(t, g.IsBlocked("1.1.1.1"))
}

func TestBruteForce_BlocksAfterMaxFailures(t *testing.T) {
	g := newTestBF(3)
	defer g.Stop()
	ip := "2.2.2.2"
	g.RecordFailure(ip)
	g.RecordFailure(ip)
	blocked := g.RecordFailure(ip)
	testkit.AssertTrue(t, blocked)
	testkit.AssertTrue(t, g.IsBlocked(ip))
}

func TestBruteForce_ResetClearsBlock(t *testing.T) {
	g := newTestBF(2)
	defer g.Stop()
	ip := "3.3.3.3"
	g.RecordFailure(ip)
	g.RecordFailure(ip)
	g.Reset(ip)
	testkit.AssertFalse(t, g.IsBlocked(ip))
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
	testkit.AssertFalse(t, g.IsBlocked(ip))
}

func TestBruteForce_IndependentIPs(t *testing.T) {
	g := newTestBF(2)
	defer g.Stop()
	g.RecordFailure("5.5.5.5")
	g.RecordFailure("5.5.5.5")
	testkit.AssertTrue(t, g.IsBlocked("5.5.5.5"))
	testkit.AssertFalse(t, g.IsBlocked("6.6.6.6"))
}
