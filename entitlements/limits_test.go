package entitlements

import (
	"encoding/json"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestGetLimitsForTier_AllTiers(t *testing.T) {
	tiers := []string{"free", "starter", "pro", "power", "concierge"}
	for _, tier := range tiers {
		limits := GetLimitsForTier(tier)
		testkit.AssertEqual(t, limits.Tier, tier)
	}
}

func TestGetLimitsForTier_UnknownFallsToFree(t *testing.T) {
	limits := GetLimitsForTier("enterprise")
	testkit.AssertEqual(t, limits.Tier, "free")
}

func TestGetLimitsForTier_EmptyFallsToFree(t *testing.T) {
	limits := GetLimitsForTier("")
	testkit.AssertEqual(t, limits.Tier, "free")
}

func TestGetLimitsForTier_FreeRestrictions(t *testing.T) {
	l := GetLimitsForTier("free")
	testkit.AssertFalse(t, l.AutoSubmit)
	testkit.AssertEqual(t, l.MonthlyApplications, 10)
	testkit.AssertEqual(t, l.ResumeVersions, 1)
	testkit.AssertEqual(t, l.TracksLimit, 1)
}

func TestGetLimitsForTier_ProFeatures(t *testing.T) {
	l := GetLimitsForTier("pro")
	testkit.AssertTrue(t, l.AutoSubmit)
	testkit.AssertTrue(t, l.SIAFIMode)
	testkit.AssertTrue(t, l.CoverLetterGen)
	testkit.AssertEqual(t, l.MonthlyApplications, 150)
}

func TestGetLimitsForTier_PowerFeatures(t *testing.T) {
	l := GetLimitsForTier("power")
	testkit.AssertTrue(t, l.AdvancedAnalytics)
	testkit.AssertTrue(t, l.PriorityQueue)
	testkit.AssertEqual(t, l.MonthlyApplications, 400)
}

func TestGetLimitsForTier_ConciergeCoaching(t *testing.T) {
	l := GetLimitsForTier("concierge")
	testkit.AssertEqual(t, l.CoachingSessions, 1)
}

func TestGetLimitsForTier_StarterGmail(t *testing.T) {
	l := GetLimitsForTier("starter")
	testkit.AssertTrue(t, l.GmailIntegration)
	testkit.AssertFalse(t, l.AutoSubmit)
}

func TestGetLimitsForTier_CaseSensitive(t *testing.T) {
	// Tier names are case-sensitive; "Free" should fall back to free defaults.
	limits := GetLimitsForTier("Free")
	testkit.AssertEqual(t, limits.Tier, "free")
}

func TestTierLimits_JSONRoundTrip(t *testing.T) {
	limits := GetLimitsForTier("pro")
	data, err := json.Marshal(limits)
	testkit.RequireNoError(t, err)
	var got TierLimits
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	testkit.AssertEqual(t, got.Tier, "pro")
	testkit.AssertEqual(t, got.MonthlyApplications, 150)
	testkit.AssertTrue(t, got.AutoSubmit)
}

func TestTierConfig_AllTiersPresent(t *testing.T) {
	expected := []string{"free", "starter", "pro", "power", "concierge"}
	for _, tier := range expected {
		if _, ok := TierConfig[tier]; !ok {
			t.Errorf("TierConfig missing tier %q", tier)
		}
	}
	testkit.AssertLen(t, TierConfig, len(expected))
}

func TestGetLimitsForTier_ProgressiveFeatures(t *testing.T) {
	free := GetLimitsForTier("free")
	starter := GetLimitsForTier("starter")
	pro := GetLimitsForTier("pro")
	power := GetLimitsForTier("power")

	testkit.AssertTrue(t, free.MonthlyApplications < starter.MonthlyApplications)
	testkit.AssertTrue(t, starter.MonthlyApplications < pro.MonthlyApplications)
	testkit.AssertTrue(t, pro.MonthlyApplications < power.MonthlyApplications)
}

func BenchmarkGetLimitsForTier_Known(b *testing.B) {
	for b.Loop() {
		GetLimitsForTier("pro")
	}
}

func BenchmarkGetLimitsForTier_Unknown(b *testing.B) {
	for b.Loop() {
		GetLimitsForTier("enterprise")
	}
}

func TestCanApplyThisMonth_BelowLimit(t *testing.T) {
	testkit.AssertTrue(t, CanApplyThisMonth("free", 9))
}

func TestCanApplyThisMonth_AtLimit(t *testing.T) {
	testkit.AssertFalse(t, CanApplyThisMonth("free", 10))
}

func TestCanApplyThisMonth_AboveLimit(t *testing.T) {
	testkit.AssertFalse(t, CanApplyThisMonth("free", 11))
}

func TestCanApplyThisMonth_ZeroCount(t *testing.T) {
	testkit.AssertTrue(t, CanApplyThisMonth("pro", 0))
}

func TestCanApplyThisMonth_UnknownTierFallsToFree(t *testing.T) {
	testkit.AssertTrue(t, CanApplyThisMonth("enterprise", 5))
	testkit.AssertFalse(t, CanApplyThisMonth("enterprise", 10))
}

func TestCanApplyThisMonth_AllTiers(t *testing.T) {
	for _, tier := range AllTierNames() {
		limits := GetLimitsForTier(tier)
		testkit.AssertTrue(t, CanApplyThisMonth(tier, limits.MonthlyApplications-1))
		testkit.AssertFalse(t, CanApplyThisMonth(tier, limits.MonthlyApplications))
	}
}

func TestAllTierNames_ContainsExpected(t *testing.T) {
	names := AllTierNames()
	expected := map[string]bool{
		"free": true, "starter": true, "pro": true,
		"power": true, "concierge": true,
	}
	testkit.AssertLen(t, names, len(expected))
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected tier %q in AllTierNames()", n)
		}
	}
}

func TestAllTierNames_Order(t *testing.T) {
	names := AllTierNames()
	want := []string{"free", "starter", "pro", "power", "concierge"}
	for i, n := range names {
		testkit.AssertEqual(t, n, want[i])
	}
}

func TestAllTierNames_IsACopy(t *testing.T) {
	names := AllTierNames()
	names[0] = "mutated"
	fresh := AllTierNames()
	testkit.AssertEqual(t, fresh[0], "free")
}

func BenchmarkAllTierNames(b *testing.B) {
	for b.Loop() {
		AllTierNames()
	}
}

func BenchmarkCanApplyThisMonth(b *testing.B) {
	for b.Loop() {
		CanApplyThisMonth("pro", 75)
	}
}

func BenchmarkIsTierValid(b *testing.B) {
	for b.Loop() {
		IsTierValid("pro")
	}
}

func TestIsTierValid_KnownTiers(t *testing.T) {
	for _, tier := range AllTierNames() {
		testkit.AssertTrue(t, IsTierValid(tier))
	}
}

func TestIsTierValid_UnknownTier(t *testing.T) {
	for _, tier := range []string{"", "enterprise", "gold", "ADMIN", "Free"} {
		testkit.AssertFalse(t, IsTierValid(tier))
	}
}
