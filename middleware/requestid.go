package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestIDKey is the context key for the request ID.
const RequestIDKey contextKey = "requestId"

// RequestIDHeader is the HTTP header used to propagate request IDs.
const RequestIDHeader = "X-Request-ID"

// GetRequestID extracts the request ID from the request context.
func GetRequestID(r *http.Request) string {
	v, _ := r.Context().Value(RequestIDKey).(string)
	return v
}

// SetRequestID returns a new context with the given request ID set.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

// RequestID is middleware that ensures every request has an X-Request-ID.
// If the incoming request already has the header, it is reused; otherwise
// a new random hex ID is generated. The ID is set on the response header
// and stored in the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = generateID()
		}

		w.Header().Set(RequestIDHeader, id)
		ctx := SetRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateID creates a random 16-byte hex string (32 chars).
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
