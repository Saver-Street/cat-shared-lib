package entitlements

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Querier is the minimal interface for database query operations.
// *pgxpool.Pool, *pgx.Conn, and pgx.Tx all satisfy this interface.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetUserTierAndUsage returns the subscription tier, monthly application count,
// and any error for the given user. If subscription_status is "past_due", returns
// "free" tier limits to enforce payment.
func GetUserTierAndUsage(ctx context.Context, db Querier, userID string) (string, int, error) {
	var tier string
	var status *string
	err := db.QueryRow(ctx,
		`SELECT COALESCE(subscription_tier, 'free'), subscription_status FROM users WHERE id = $1`,
		userID,
	).Scan(&tier, &status)
	if err != nil {
		return "free", 0, err
	}

	// past_due users get free tier limits until payment resolves
	if status != nil && *status == "past_due" {
		tier = "free"
	}

	var appCount int
	_ = db.QueryRow(ctx,
		`SELECT COUNT(*) FROM applications a
		 JOIN candidate_profiles cp ON cp.id = a.candidate_id
		 WHERE cp.user_id = $1
		 AND a.created_date >= date_trunc('month', CURRENT_TIMESTAMP)`,
		userID,
	).Scan(&appCount)

	return tier, appCount, nil
}

// GetUserTier is a convenience wrapper that returns only the tier string.
func GetUserTier(ctx context.Context, db Querier, userID string) (string, error) {
	tier, _, err := GetUserTierAndUsage(ctx, db, userID)
	return tier, err
}
