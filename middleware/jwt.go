package middleware

import (
	"context"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// contextKey is the package-private type for context keys.
type contextKey string

const (
	// UserIDKey is the context key for the authenticated user ID.
	UserIDKey contextKey = "userId"
	// UserRoleKey is the context key for the user role.
	UserRoleKey contextKey = "userRole"
	// UserEmailKey is the context key for the user email.
	UserEmailKey contextKey = "userEmail"
	// ExtCandidateIDKey is the context key set by extension token middleware.
	ExtCandidateIDKey contextKey = "extCandidateId"
	// ExtUserIDKey is the context key for the extension user ID.
	ExtUserIDKey contextKey = "extUserId"
	// ExtTokenIDKey is the context key for the extension token ID.
	ExtTokenIDKey contextKey = "extTokenId"
)

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(UserIDKey).(string)
	return v
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(r *http.Request) string {
	v, _ := r.Context().Value(UserRoleKey).(string)
	return v
}

// GetUserEmail extracts the user email from the request context.
func GetUserEmail(r *http.Request) string {
	v, _ := r.Context().Value(UserEmailKey).(string)
	return v
}

// GetExtCandidateID extracts the extension-provided candidate ID from context.
func GetExtCandidateID(r *http.Request) string {
	v, _ := r.Context().Value(ExtCandidateIDKey).(string)
	return v
}

// SetUserID returns a new context with the given user ID set.
// Used by each service's own JWT parsing middleware.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SetUserRole returns a new context with the given user role set.
func SetUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

// SetUserEmail returns a new context with the given user email set.
func SetUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, UserEmailKey, email)
}

// SetExtCandidateID returns a new context with the extension candidate ID set.
func SetExtCandidateID(ctx context.Context, candidateID string) context.Context {
	return context.WithValue(ctx, ExtCandidateIDKey, candidateID)
}

// GetExtUserID extracts the extension user ID from the request context.
func GetExtUserID(r *http.Request) string {
	v, _ := r.Context().Value(ExtUserIDKey).(string)
	return v
}

// SetExtUserID returns a new context with the extension user ID set.
func SetExtUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ExtUserIDKey, userID)
}

// GetExtTokenID extracts the extension token ID from the request context.
func GetExtTokenID(r *http.Request) string {
	v, _ := r.Context().Value(ExtTokenIDKey).(string)
	return v
}

// SetExtTokenID returns a new context with the extension token ID set.
func SetExtTokenID(ctx context.Context, tokenID string) context.Context {
	return context.WithValue(ctx, ExtTokenIDKey, tokenID)
}

// RequireAuth is a middleware that rejects unauthenticated requests.
// It expects the user ID to have been set in context by an upstream auth middleware.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserID(r) == "" {
			response.Error(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin rejects requests where the user role is not "admin".
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserID(r) == "" {
			response.Error(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		if GetUserRole(r) != "admin" {
			response.Error(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole returns a middleware that allows only users with the given role.
// Returns 401 if the request has no authenticated user, 403 if the role does not match.
func RequireRole(role string) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if GetUserID(r) == "" {
response.Error(w, http.StatusUnauthorized, "Authentication required")
return
}
if GetUserRole(r) != role {
response.Error(w, http.StatusForbidden, "Insufficient role")
return
}
next.ServeHTTP(w, r)
})
}
}
