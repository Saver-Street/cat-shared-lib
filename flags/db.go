package flags

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// Querier is the minimal interface for database query operations.
// *pgxpool.Pool, *pgx.Conn, and pgx.Tx all satisfy this interface.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// queryFlag looks up a flag key in site_settings and returns its value.
// Returns defaultValue if the row is absent or db is nil.
// Logs and returns defaultValue on any other database error.
func queryFlag(ctx context.Context, db Querier, key string, defaultValue bool) bool {
	if db == nil {
		return defaultValue
	}
	var value string
	err := db.QueryRow(ctx,
		"SELECT value FROM site_settings WHERE key = $1", key,
	).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return defaultValue
	}
	if err != nil {
		slog.Error("flags: failed to query flag", "key", key, "error", err)
		return defaultValue
	}
	return value == "true"
}

// IsFeatureEnabled checks if a named feature flag is enabled.
// Boolean flags are stored as the literal string "true" in site_settings
// (no encryption — per cat-shared-lib design: boolean flags are plain-text).
// Returns true if the flag row is absent (safe default = enabled).
func IsFeatureEnabled(ctx context.Context, db Querier, flagName string) bool {
	return queryFlag(ctx, db, "flag_"+flagName, true)
}

// IsMaintenanceModeActive returns true only when flag_maintenanceMode is
// explicitly set to "true" in site_settings. Defaults to false.
func IsMaintenanceModeActive(ctx context.Context, db Querier) bool {
	return queryFlag(ctx, db, "flag_"+FlagMaintenanceMode, false)
}

// IsGlobalAutomationPaused returns true when flag_globalAutomationPause is "true".
// Defaults to false (automation runs normally).
func IsGlobalAutomationPaused(ctx context.Context, db Querier) bool {
	return queryFlag(ctx, db, "flag_"+FlagGlobalAutoPause, false)
}
