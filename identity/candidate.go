package identity

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKey is the package-private type for context keys to avoid collisions.
type contextKey string

const (
	userIDKey         contextKey = "userId"
	extCandidateIDKey contextKey = "extCandidateId"
)

// GetUserID extracts the authenticated user ID from the request context.
// Returns empty string if not set.
func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(userIDKey).(string)
	return v
}

// GetExtCandidateID extracts the extension-provided candidate ID from context.
// Returns empty string if not set (only present for extension token requests).
func GetExtCandidateID(r *http.Request) string {
	v, _ := r.Context().Value(extCandidateIDKey).(string)
	return v
}

// LookupCandidateID queries candidate_profiles for the candidate ID of the given user.
// Returns the candidate ID, or an error if no profile exists.
func LookupCandidateID(ctx context.Context, pool *pgxpool.Pool, userID string) (string, error) {
	var candidateID string
	err := pool.QueryRow(ctx,
		"SELECT id FROM candidate_profiles WHERE user_id = $1", userID,
	).Scan(&candidateID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("candidate profile not found for user " + userID)
	}
	return candidateID, err
}

// ResolveCandidate returns the candidate ID for the current request.
// Checks for an extension-provided candidate ID first, then falls back to
// looking up the candidate profile for the authenticated user.
// Returns empty string (not an error) if no identity is present.
func ResolveCandidate(r *http.Request, pool *pgxpool.Pool) (string, error) {
	if id := GetExtCandidateID(r); id != "" {
		return id, nil
	}
	userID := GetUserID(r)
	if userID == "" {
		return "", nil
	}
	return LookupCandidateID(r.Context(), pool, userID)
}
