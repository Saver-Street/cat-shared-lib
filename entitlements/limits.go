// Package entitlements provides subscription tier lookups and usage checks.
package entitlements

// GetLimitsForTier returns the TierLimits for the given tier name.
// Falls back to free tier limits for unknown tiers.
func GetLimitsForTier(tier string) TierLimits {
	if limits, ok := TierConfig[tier]; ok {
		return limits
	}
	return TierConfig["free"]
}

// tierOrder defines the canonical ordering of subscription tiers from least to most capable.
var tierOrder = []string{"free", "starter", "pro", "power", "concierge"}

// AllTierNames returns the canonical list of subscription tier names in ascending capability order.
// The returned slice is a copy; callers may modify it freely.
func AllTierNames() []string {
	out := make([]string, len(tierOrder))
	copy(out, tierOrder)
	return out
}

// CanApplyThisMonth reports whether a user on the given tier may submit another
// job application this calendar month, given they have already submitted count applications.
// Falls back to free-tier limits for unknown tier names.
func CanApplyThisMonth(tier string, count int) bool {
	return count < GetLimitsForTier(tier).MonthlyApplications
}

// IsTierValid reports whether tier is a known subscription tier.
func IsTierValid(tier string) bool {
	_, ok := TierConfig[tier]
	return ok
}
