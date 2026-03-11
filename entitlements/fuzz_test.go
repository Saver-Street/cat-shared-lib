package entitlements

import "testing"

func FuzzGetLimitsForTier(f *testing.F) {
	f.Add("free")
	f.Add("starter")
	f.Add("pro")
	f.Add("power")
	f.Add("concierge")
	f.Add("unknown")
	f.Add("")
	f.Add("FREE")
	f.Add("Pro")
	f.Fuzz(func(t *testing.T, tier string) {
		limits := GetLimitsForTier(tier)
		// Must always return valid limits (defaults to free).
		if limits.MonthlyApplications < 0 {
			t.Error("MonthlyApplications is negative")
		}
	})
}

func FuzzCanApplyThisMonth(f *testing.F) {
	f.Add("free", 0)
	f.Add("pro", 100)
	f.Add("unknown", -1)
	f.Add("", 0)
	f.Add("concierge", 999999)
	f.Fuzz(func(t *testing.T, tier string, count int) {
		// Must not panic on any input.
		_ = CanApplyThisMonth(tier, count)
	})
}

func FuzzIsTierValid(f *testing.F) {
	f.Add("free")
	f.Add("pro")
	f.Add("invalid")
	f.Add("")
	f.Add("FREE")
	f.Fuzz(func(t *testing.T, tier string) {
		valid := IsTierValid(tier)
		// If valid, GetLimitsForTier should return matching config.
		if valid {
			limits := GetLimitsForTier(tier)
			if limits.MonthlyApplications <= 0 {
				t.Error("valid tier returned zero/negative monthly applications")
			}
		}
	})
}
