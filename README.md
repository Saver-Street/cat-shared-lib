# cat-shared-lib

[![CI](https://github.com/Saver-Street/cat-shared-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/Saver-Street/cat-shared-lib/actions/workflows/ci.yml)

Shared Go library for Catherine (Auto-Apps) microservices.

**Wave 0 — required by all service extractions.**

## Packages

| Package | Description | Coverage |
|---|---|---|
| `middleware` | JWT auth (HS256), request ID, logging, recovery, rate limiting, brute-force | 99% |
| `response` | JSON response helpers, pagination headers | 100% |
| `request` | HTTP request parsing, URL param extraction, pagination | 100% |
| `health` | Standardized health check handlers with concurrent checkers | 100% |
| `server` | HTTP server with graceful shutdown (SIGINT/SIGTERM) | 100% |
| `validation` | Email, UUID, phone, URL validators with clear error messages | 98% |
| `apperror` | Standardized error types with HTTP status codes | 100% |
| `config` | Env var parsing with defaults, validation, required checks | 100% |
| `database` | Connection pool setup, migration runner, transaction helpers | 81% |
| `entitlements` | Subscription tier limits + DirectDB queries | 100% |
| `flags` | Feature flag DirectDB reads (boolean flags, plain-text) | 100% |
| `identity` | Candidate resolution + context getters | 100% |
| `types` | Shared domain types (User, CandidateProfile, Pagination) | 100% |
| `scan` | Generic database row scanning (Rows, Row, First) | 100% |
| `sanitize` | Filename sanitization, NilIfEmpty, IsDuplicateKey, Deref | 100% |
| `security` | Suspicious input detection, PII redaction | 100% |

## Usage

```go
import (
    "github.com/Saver-Street/cat-shared-lib/middleware"
    "github.com/Saver-Street/cat-shared-lib/response"
    "github.com/Saver-Street/cat-shared-lib/request"
    "github.com/Saver-Street/cat-shared-lib/health"
    "github.com/Saver-Street/cat-shared-lib/server"
    "github.com/Saver-Street/cat-shared-lib/validation"
    "github.com/Saver-Street/cat-shared-lib/apperror"
    "github.com/Saver-Street/cat-shared-lib/config"
    "github.com/Saver-Street/cat-shared-lib/database"
    "github.com/Saver-Street/cat-shared-lib/entitlements"
    "github.com/Saver-Street/cat-shared-lib/flags"
    "github.com/Saver-Street/cat-shared-lib/identity"
    "github.com/Saver-Street/cat-shared-lib/scan"
    "github.com/Saver-Street/cat-shared-lib/sanitize"
    "github.com/Saver-Street/cat-shared-lib/security"
)
```

## Package Examples

### middleware — JWT Auth (HS256)

```go
// Create JWT auth middleware
authMW := middleware.JWTAuth(middleware.JWTAuthConfig{
    Secret:    []byte(os.Getenv("JWT_SECRET")),
    Issuer:    "cat-service",
    SkipPaths: []string{"/health", "/ready"},
})

mux := http.NewServeMux()
mux.Handle("/api/", authMW(apiHandler))

// In handlers, read authenticated user info:
userID := middleware.GetUserID(r)
email := middleware.GetUserEmail(r)
role := middleware.GetUserRole(r)

// Issue a token:
token, _ := middleware.SignHS256(middleware.JWTClaims{
    Subject:   userID,
    Email:     email,
    Role:      "admin",
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    Issuer:    "cat-service",
}, secret)
```

### middleware — Request ID, Logging, Recovery

```go
// Stack middleware (outermost first):
handler := middleware.Recovery(
    middleware.RequestID(
        middleware.Logging(logger)(
            middleware.RequireAuth(
                myHandler,
            ),
        ),
    ),
)

// Access request ID anywhere in the chain:
reqID := middleware.GetRequestID(r)
```

### middleware — Authorization

```go
// Require authentication
mux.Handle("/api/", middleware.RequireAuth(handler))

// Require specific role
mux.Handle("/admin/", middleware.RequireRole("admin")(handler))

// Require minimum subscription tier
mux.Handle("/premium/", middleware.RequireSubscriptionTier("pro")(handler))
```

### response — JSON Responses

```go
response.OK(w, user)                           // 200
response.Created(w, newRecord)                  // 201
response.NoContent(w)                           // 204
response.BadRequest(w, "invalid email")         // 400
response.NotFound(w, "user not found")          // 404
response.InternalError(w, "fetch failed", err)  // 500 (logs error)

// Paginated with headers
response.PaginatedWithHeaders(w, users, total, page, limit)
// Sets X-Total-Count, X-Limit, X-Offset headers + JSON body

// Decode request body
var req CreateUserRequest
if !response.DecodeOrFail(w, r, &req) {
    return // 400 already sent
}
```

### request — Parsing

```go
// Parse pagination from query params
p := request.ParsePagination(r.URL.Query(), 20, 100)
// p.Page, p.Limit, p.Offset

// Required URL params
id, err := request.RequireURLParamInt(r, "id", chi.URLParam)

// Optional query params
status := request.OptionalQueryParam(r.URL.Query(), "status", "active")
```

### health — Health Checks

```go
healthHandler := health.Handler("billing-service", "1.2.0",
    health.DBChecker(pool),
    health.NewChecker("redis", func(ctx context.Context) error {
        return redisClient.Ping(ctx).Err()
    }),
)
mux.Handle("/health", healthHandler)
// Returns: {"status":"ok","service":"billing-service","version":"1.2.0","uptime":"5m30s","checks":{"db":"ok","redis":"ok"}}
```

### validation — Input Validation

```go
errs := validation.Collect(
    validation.Email("email", req.Email),
    validation.UUID("userId", req.UserID),
    validation.Phone("phone", req.Phone),
    validation.URL("website", req.Website),
    validation.Required("name", req.Name),
    validation.MinLength("password", req.Password, 8),
    validation.OneOf("role", req.Role, []string{"admin", "user", "guest"}),
)
if errs != nil {
    // errs is []error, each is *validation.ValidationError with Field and Message
    response.BadRequest(w, errs[0].Error())
    return
}
```

### apperror — Structured Errors

```go
// Create typed errors
err := apperror.NotFound("user not found")
err := apperror.BadRequest("invalid email format")
err := apperror.InternalWrap("query failed", dbErr)

// Check error type
if apperror.IsCode(err, apperror.CodeNotFound) {
    response.NotFound(w, err.Error())
    return
}

// Get HTTP status from any error
status := apperror.HTTPStatus(err) // 404, 400, 500, etc.
```

### config — Environment Variables

```go
port := config.Int("PORT", 8080)
dbURL := config.MustString("DATABASE_URL")  // panics if unset
debug := config.Bool("DEBUG", false)
timeout := config.Duration("TIMEOUT", 30*time.Second)
origins := config.StringSlice("CORS_ORIGINS", []string{"http://localhost:3000"})

// Validate all required vars at startup
if err := config.Validate("DATABASE_URL", "JWT_SECRET", "REDIS_URL"); err != nil {
    log.Fatal(err) // "config: missing required environment variables: JWT_SECRET, REDIS_URL"
}
```

### database — Pool, Transactions, Migrations

```go
// Create connection pool
pool, err := database.NewPool(ctx, database.PoolConfig{
    DSN:      config.MustString("DATABASE_URL"),
    MaxConns: 20,
    MinConns: 5,
})
defer pool.Close()

// Run migrations
err = database.Migrate(ctx, pool, []database.Migration{
    {Version: 1, Name: "create_users", SQL: "CREATE TABLE users (id UUID PRIMARY KEY, email TEXT NOT NULL)"},
    {Version: 2, Name: "add_role", SQL: "ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user'"},
})

// Transaction helper
err = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, email)
    if err != nil {
        return err // automatically rolled back
    }
    _, err = tx.Exec(ctx, "INSERT INTO profiles (user_id) VALUES ($1)", id)
    return err // committed if nil
})
```

### scan — Generic Row Scanning

```go
type User struct { ID string; Email string }

users, err := scan.Rows[User](rows, func(u *User) []any {
    return []any{&u.ID, &u.Email}
})

user, err := scan.Row[User](pool.QueryRow(ctx, sql, id), func(u *User) []any {
    return []any{&u.ID, &u.Email}
})
```

## Querier Interface

DB-querying functions accept a `Querier` interface instead of concrete `*pgxpool.Pool`:

```go
type Querier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Both `*pgxpool.Pool`, `*pgx.Conn`, and `pgx.Tx` satisfy this interface, making the
library flexible and fully testable without a real database.

## Design Notes

- **DirectDB only** in Phase C — no HTTP service-to-service calls
- **Boolean flags are plain-text** (`"true"` / `"false"`) — no encryption
- **Rate limiter** is per-IP sliding window + token bucket, safe for concurrent use
- **98%+ test coverage** across all packages
- **Querier interface** for DB functions — accepts pool, conn, or tx

## Import in services

```go
// go.mod
require github.com/Saver-Street/cat-shared-lib v1.5.0
```
