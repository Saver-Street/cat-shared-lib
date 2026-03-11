// Package entitlements manages subscription tier definitions, feature limits,
// and usage tracking for the Catherine platform.
//
// [TierConfig] maps each tier name (free, starter, pro, power, concierge) to a
// [TierLimits] struct that specifies caps on monthly applications, resume
// versions, tracked jobs, and boolean feature gates such as AutoSubmit and
// GmailIntegration.  Use [GetLimitsForTier] to look up limits by tier name
// and [CanApplyThisMonth] to check whether a user's usage is within quota.
//
// [GetUserTierAndUsage] queries the database for a user's current tier and
// monthly application count, automatically downgrading past-due accounts to
// the free tier.  [AllTierNames] and [IsTierValid] help with validation.
package entitlements
