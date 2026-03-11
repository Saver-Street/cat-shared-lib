package entitlements

import (
	"context"
	"sync"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestGetUserTierAndUsage_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	db := &mockQuerier{calls: []mockCall{
		{row: &mockRow{scanFunc: func(_ ...any) error {
			return ctx.Err()
		}}},
	}}

	tier, count, err := GetUserTierAndUsage(ctx, db, "user-cancel")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	testkit.AssertEqual(t, tier, "free")
	testkit.AssertEqual(t, count, 0)
}

func TestGetUserTierAndUsage_Concurrent(t *testing.T) {
	active := "active"
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			db := &mockQuerier{calls: []mockCall{
				{row: &mockRow{scanFunc: func(dest ...any) error {
					*dest[0].(*string) = "pro"
					*dest[1].(**string) = &active
					return nil
				}}},
				{row: &mockRow{scanFunc: func(dest ...any) error {
					*dest[0].(*int) = 10
					return nil
				}}},
			}}
			tier, count, err := GetUserTierAndUsage(context.Background(), db, "user-concurrent")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tier != "pro" {
				t.Errorf("tier = %q, want pro", tier)
			}
			if count != 10 {
				t.Errorf("count = %d, want 10", count)
			}
		}()
	}
	wg.Wait()
}

func FuzzGetUserTier(f *testing.F) {
	f.Add("user-123")
	f.Add("")
	f.Add("a")
	f.Add("user with spaces")

	f.Fuzz(func(t *testing.T, userID string) {
		active := "active"
		db := &mockQuerier{calls: []mockCall{
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*string) = "pro"
				*dest[1].(**string) = &active
				return nil
			}}},
			{row: &mockRow{scanFunc: func(dest ...any) error {
				*dest[0].(*int) = 5
				return nil
			}}},
		}}
		// Must not panic
		tier, err := GetUserTier(context.Background(), db, userID)
		testkit.RequireNoError(t, err)
		testkit.AssertEqual(t, tier, "pro")
	})
}

func TestTierConfig_AllTiersHaveConsistentLimits(t *testing.T) {
	// Verify tier hierarchy: higher tiers must have >= limits than lower ones
	tiers := []string{"free", "starter", "pro", "power", "concierge"}
	for i, name := range tiers {
		limits, ok := TierConfig[name]
		if !ok {
			t.Fatalf("TierConfig missing tier %q", name)
		}
		testkit.AssertEqual(t, limits.Tier, name)
		if i > 0 {
			prev := TierConfig[tiers[i-1]]
			if limits.MonthlyApplications < prev.MonthlyApplications {
				t.Errorf("%s apps (%d) < %s apps (%d)", name, limits.MonthlyApplications, tiers[i-1], prev.MonthlyApplications)
			}
		}
	}
}

func TestTierConfig_ConciergeHasCoaching(t *testing.T) {
	c := TierConfig["concierge"]
	if c.CoachingSessions < 1 {
		t.Errorf("concierge should have coaching sessions, got %d", c.CoachingSessions)
	}
	// No other tier should have coaching
	for name, limits := range TierConfig {
		if name != "concierge" && limits.CoachingSessions > 0 {
			t.Errorf("tier %q unexpectedly has %d coaching sessions", name, limits.CoachingSessions)
		}
	}
}
