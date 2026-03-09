package flags

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Querier is the minimal interface for database query operations.
// *pgxpool.Pool, *pgx.Conn, and pgx.Tx all satisfy this interface.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// IsFeatureEnabled checks if a named feature flag is enabled.
// Boolean flags are stored as the literal string "true" in site_settings
// (no encryption — per cat-shared-lib design: boolean flags are plain-text).
// Returns true if the flag row is absent (safe default = enabled).
func IsFeatureEnabled(ctx context.Context, db Querier, flagName string) bool {
	if db == nil {
		return true
	}
	var value string
	err := db.QueryRow(ctx,
		"SELECT value FROM site_settings WHERE key = $1", "flag_"+flagName,
	).Scan(&value)
	if err == pgx.ErrNoRows {
		return true
	}
	if err != nil {
		return true
	}
	return value == "true"
}

// IsMaintenanceModeActive returns true only when flag_maintenanceMode is
// explicitly set to "true" in site_settings. Defaults to false.
func IsMaintenanceModeActive(ctx context.Context, db Querier) bool {
	if db == nil {
		return false
	}
	var value string
	err := db.QueryRow(ctx,
		"SELECT value FROM site_settings WHERE key = 'flag_maintenanceMode'",
	).Scan(&value)
	if err != nil {
		return false
	}
	return value == "true"
}

// IsGlobalAutomationPaused returns true when flag_globalAutomationPause is "true".
// Defaults to false (automation runs normally).
func IsGlobalAutomationPaused(ctx context.Context, db Querier) bool {
	if db == nil {
		return false
	}
	var value string
	err := db.QueryRow(ctx,
		"SELECT value FROM site_settings WHERE key = 'flag_globalAutomationPause'",
	).Scan(&value)
	if err != nil {
		return false
	}
	return value == "true"
}
