// Package flags provides database-backed feature flag queries against a
// site_settings table, complementing the environment-variable approach in the
// featureflags package.
//
// Predefined flag constants such as [FlagAIScoring], [FlagMaintenanceMode],
// and [FlagGlobalAutoPause] identify well-known operational toggles.
// [IsFeatureEnabled] checks whether a flag is enabled (defaulting to true for
// feature flags), while [IsMaintenanceModeActive] and
// [IsGlobalAutomationPaused] use safe defaults of false.
// [IsCustomFlagEnabled] allows querying arbitrary keys with a caller-supplied
// default.
//
// All functions accept a [Querier] interface so they work with any pgx-
// compatible query executor (pool, connection, or transaction).
package flags
