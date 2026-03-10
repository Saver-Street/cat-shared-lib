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
