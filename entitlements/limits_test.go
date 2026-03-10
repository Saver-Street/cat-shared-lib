package entitlements

import "testing"

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
