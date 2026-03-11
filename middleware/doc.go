// Package middleware provides HTTP middleware for authentication,
// authorization, rate limiting, observability, and resilience in Catherine
// microservices.
//
// # Authentication & Authorization
//
// [JWTAuth] validates HS256 JSON Web Tokens and populates the request context
// with user identity fields accessible via [GetUserID], [GetUserRole],
// [GetUserEmail], and related getters.  [RequireAuth], [RequireAdmin],
// [RequireRole], and [RequireSubscriptionTier] enforce access policies.
// [SignHS256] creates signed tokens for testing or token issuance.
//
// [CORS] handles Cross-Origin Resource Sharing with configurable origins,
// methods, headers, credentials, and max-age via [CORSConfig].
//
// # Rate Limiting & Brute-Force Protection
//
// [NewRateLimiter] implements a sliding-window rate limiter keyed by client IP.
// [NewTokenBucketLimiter] offers token-bucket semantics.
// [NewBruteForceGuard] blocks IPs after repeated failures.
//
// # Observability & Resilience
//
// [Logging] logs request method, path, status, and duration.  [RequestID]
// injects a unique request ID into the context and response headers.
// [Recovery] catches panics and returns 500.  [Timeout] enforces a per-request
// deadline.
//
// # Composition
//
// [Chain] composes multiple middleware into a single wrapper, applying them
// in order so that the first argument is the outermost layer.
package middleware
