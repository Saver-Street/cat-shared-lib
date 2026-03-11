package middleware

import (
	"context"
	"net/http"
)

// CorrelationIDKey is the context key for the correlation ID.
const CorrelationIDKey contextKey = "correlationId"

// CorrelationIDHeader is the HTTP header used to propagate correlation IDs
// across service boundaries.
const CorrelationIDHeader = "X-Correlation-ID"

// GetCorrelationID extracts the correlation ID from the request context.
func GetCorrelationID(r *http.Request) string {
	v, _ := r.Context().Value(CorrelationIDKey).(string)
	return v
}

// CorrelationIDFromContext extracts the correlation ID directly from a context.
func CorrelationIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CorrelationIDKey).(string)
	return v
}

// SetCorrelationID returns a new context with the given correlation ID set.
func SetCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// CorrelationID is middleware that ensures every request carries a correlation ID
// for distributed tracing across service boundaries. If the incoming request has
// an X-Correlation-ID header (within the allowed length), it is reused to maintain
// traceability across services. Otherwise, a new random hex ID is generated.
//
// The correlation ID is set on the response header and stored in the request context.
// Downstream services should forward the X-Correlation-ID header to maintain the
// trace chain.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(CorrelationIDHeader)
		if id == "" || len(id) > maxRequestIDLen {
			id = generateID()
		}

		w.Header().Set(CorrelationIDHeader, id)
		ctx := SetCorrelationID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
