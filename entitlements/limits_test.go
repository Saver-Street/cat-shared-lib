package entitlements

import (
	"encoding/json"
	"testing"
)

func TestGetLimitsForTier_AllTiers(t *testing.T) {
	tiers := []string{"free", "starter", "pro", "power", "concierge"}
	for _, tier := range tiers {
		limits := GetLimitsForTier(tier)
		if limits.Tier != tier {
			t.Errorf("tier %q: Tier field = %q, want %q", tier, limits.Tier, tier)
		}
	}
}

func TestGetLimitsForTier_UnknownFallsToFree(t *testing.T) {
	limits := GetLimitsForTier("enterprise")
	if limits.Tier != "free" {
		t.Errorf("unknown tier should return free, got %q", limits.Tier)
	}
}

func TestGetLimitsForTier_EmptyFallsToFree(t *testing.T) {
	limits := GetLimitsForTier("")
	if limits.Tier != "free" {
		t.Errorf("empty tier should return free, got %q", limits.Tier)
	}
}

func TestGetLimitsForTier_FreeRestrictions(t *testing.T) {
	l := GetLimitsForTier("free")
	if l.AutoSubmit {
		t.Error("free: AutoSubmit should be false")
	}
	if l.MonthlyApplications != 10 {
		t.Errorf("free: MonthlyApplications = %d, want 10", l.MonthlyApplications)
	}
	if l.ResumeVersions != 1 {
		t.Errorf("free: ResumeVersions = %d, want 1", l.ResumeVersions)
	}
	if l.TracksLimit != 1 {
		t.Errorf("free: TracksLimit = %d, want 1", l.TracksLimit)
	}
}

func TestGetLimitsForTier_ProFeatures(t *testing.T) {
	l := GetLimitsForTier("pro")
	if !l.AutoSubmit {
		t.Error("pro: AutoSubmit should be true")
	}
	if !l.SIAFIMode {
		t.Error("pro: SIAFIMode should be true")
	}
	if !l.CoverLetterGen {
		t.Error("pro: CoverLetterGen should be true")
	}
	if l.MonthlyApplications != 150 {
		t.Errorf("pro: MonthlyApplications = %d, want 150", l.MonthlyApplications)
	}
}

func TestGetLimitsForTier_PowerFeatures(t *testing.T) {
	l := GetLimitsForTier("power")
	if !l.AdvancedAnalytics {
		t.Error("power: AdvancedAnalytics should be true")
	}
	if !l.PriorityQueue {
		t.Error("power: PriorityQueue should be true")
	}
	if l.MonthlyApplications != 400 {
		t.Errorf("power: MonthlyApplications = %d, want 400", l.MonthlyApplications)
	}
}

func TestGetLimitsForTier_ConciergeCoaching(t *testing.T) {
	l := GetLimitsForTier("concierge")
	if l.CoachingSessions != 1 {
		t.Errorf("concierge: CoachingSessions = %d, want 1", l.CoachingSessions)
	}
}

func TestGetLimitsForTier_StarterGmail(t *testing.T) {
	l := GetLimitsForTier("starter")
	if !l.GmailIntegration {
		t.Error("starter: GmailIntegration should be true")
	}
	if l.AutoSubmit {
		t.Error("starter: AutoSubmit should be false")
	}
}

func TestGetLimitsForTier_CaseSensitive(t *testing.T) {
	// Tier names are case-sensitive; "Free" should fall back to free defaults.
	limits := GetLimitsForTier("Free")
	if limits.Tier != "free" {
		t.Errorf("uppercase Free should fall back to free tier, got %q", limits.Tier)
	}
}

func TestTierLimits_JSONRoundTrip(t *testing.T) {
	limits := GetLimitsForTier("pro")
	data, err := json.Marshal(limits)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got TierLimits
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Tier != "pro" {
		t.Errorf("tier = %q, want pro", got.Tier)
	}
	if got.MonthlyApplications != 150 {
		t.Errorf("monthly apps = %d, want 150", got.MonthlyApplications)
	}
	if !got.AutoSubmit {
		t.Error("AutoSubmit should be true")
	}
}

func TestTierConfig_AllTiersPresent(t *testing.T) {
	expected := []string{"free", "starter", "pro", "power", "concierge"}
	for _, tier := range expected {
		if _, ok := TierConfig[tier]; !ok {
			t.Errorf("TierConfig missing tier %q", tier)
		}
	}
	if len(TierConfig) != len(expected) {
		t.Errorf("TierConfig has %d tiers, want %d", len(TierConfig), len(expected))
	}
}

func TestGetLimitsForTier_ProgressiveFeatures(t *testing.T) {
	free := GetLimitsForTier("free")
	starter := GetLimitsForTier("starter")
	pro := GetLimitsForTier("pro")
	power := GetLimitsForTier("power")

	if free.MonthlyApplications >= starter.MonthlyApplications {
		t.Error("free should have fewer apps than starter")
	}
	if starter.MonthlyApplications >= pro.MonthlyApplications {
		t.Error("starter should have fewer apps than pro")
	}
	if pro.MonthlyApplications >= power.MonthlyApplications {
		t.Error("pro should have fewer apps than power")
	}
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
if !CanApplyThisMonth("free", 9) {
t.Error("9/10 apps: should still be allowed")
}
}

func TestCanApplyThisMonth_AtLimit(t *testing.T) {
if CanApplyThisMonth("free", 10) {
t.Error("10/10 apps: should be blocked (at limit)")
}
}

func TestCanApplyThisMonth_AboveLimit(t *testing.T) {
if CanApplyThisMonth("free", 11) {
t.Error("11/10 apps: should be blocked (over limit)")
}
}

func TestCanApplyThisMonth_ZeroCount(t *testing.T) {
if !CanApplyThisMonth("pro", 0) {
t.Error("0 apps on pro: should be allowed")
}
}

func TestCanApplyThisMonth_UnknownTierFallsToFree(t *testing.T) {
if !CanApplyThisMonth("enterprise", 5) {
t.Error("5/10 on unknown tier: should be allowed (free fallback)")
}
if CanApplyThisMonth("enterprise", 10) {
t.Error("10/10 on unknown tier: should be blocked (free fallback)")
}
}

func TestCanApplyThisMonth_AllTiers(t *testing.T) {
for _, tier := range AllTierNames() {
limits := GetLimitsForTier(tier)
if !CanApplyThisMonth(tier, limits.MonthlyApplications-1) {
t.Errorf("%s: one below limit should be allowed", tier)
}
if CanApplyThisMonth(tier, limits.MonthlyApplications) {
t.Errorf("%s: at-limit should be blocked", tier)
}
}
}

func TestAllTierNames_ContainsExpected(t *testing.T) {
names := AllTierNames()
expected := map[string]bool{
"free": true, "starter": true, "pro": true,
"power": true, "concierge": true,
}
if len(names) != len(expected) {
t.Errorf("AllTierNames() len = %d, want %d", len(names), len(expected))
}
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
if n != want[i] {
t.Errorf("AllTierNames()[%d] = %q, want %q", i, n, want[i])
}
}
}

func TestAllTierNames_IsACopy(t *testing.T) {
names := AllTierNames()
names[0] = "mutated"
fresh := AllTierNames()
if fresh[0] != "free" {
t.Errorf("AllTierNames returned mutable slice; first element = %q, want free", fresh[0])
}
}

func BenchmarkCanApplyThisMonth(b *testing.B) {
for b.Loop() {
CanApplyThisMonth("pro", 75)
}
}
