// Package identity provides candidate identity resolution for HTTP handlers.
// It bridges the JWT-authenticated user ID (set by the middleware package) to
// the application's candidate profile concept.
package identity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/middleware"
	"github.com/jackc/pgx/v5"
)

// Querier is the minimal interface for database query operations.
// *pgxpool.Pool, *pgx.Conn, and pgx.Tx all satisfy this interface.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetUserID extracts the authenticated user ID from the request context.
// Returns empty string if not set. Reads the value set by middleware.SetUserID.
func GetUserID(r *http.Request) string {
	return middleware.GetUserID(r)
}

// GetExtCandidateID extracts the extension-provided candidate ID from context.
// Returns empty string if not set (only present for extension token requests).
// Reads the value set by extension token middleware via middleware.ExtCandidateIDKey.
func GetExtCandidateID(r *http.Request) string {
	return middleware.GetExtCandidateID(r)
}

// LookupCandidateID queries candidate_profiles for the candidate ID of the given user.
// Returns the candidate ID, or an error if no profile exists.
func LookupCandidateID(ctx context.Context, db Querier, userID string) (string, error) {
	var candidateID string
	err := db.QueryRow(ctx,
		"SELECT id FROM candidate_profiles WHERE user_id = $1", userID,
	).Scan(&candidateID)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("candidate profile not found for user %s", userID)
	}
	return candidateID, err
}

// ResolveCandidate returns the candidate ID for the current request.
// Checks for an extension-provided candidate ID first, then falls back to
// looking up the candidate profile for the authenticated user.
// Returns empty string (not an error) if no identity is present.
func ResolveCandidate(r *http.Request, db Querier) (string, error) {
	if id := GetExtCandidateID(r); id != "" {
		return id, nil
	}
	userID := GetUserID(r)
	if userID == "" {
		return "", nil
	}
	return LookupCandidateID(r.Context(), db, userID)
}
