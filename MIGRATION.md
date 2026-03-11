# Migration Guide

This guide helps teams adopt `cat-shared-lib` in existing Catherine
microservices or upgrade between versions.

---

## Table of Contents

- [Adopting cat-shared-lib](#adopting-cat-shared-lib)
  - [Prerequisites](#prerequisites)
  - [Step 1 — Add the Dependency](#step-1--add-the-dependency)
  - [Step 2 — Replace Inline Utilities](#step-2--replace-inline-utilities)
  - [Step 3 — Adopt the Middleware Stack](#step-3--adopt-the-middleware-stack)
  - [Step 4 — Migrate Error Handling](#step-4--migrate-error-handling)
  - [Step 5 — Adopt Configuration Helpers](#step-5--adopt-configuration-helpers)
  - [Step 6 — Switch to Shared Database Helpers](#step-6--switch-to-shared-database-helpers)
  - [Step 7 — Adopt Observability Packages](#step-7--adopt-observability-packages)
  - [Step 8 — Switch to Shared Test Helpers](#step-8--switch-to-shared-test-helpers)
- [Package-by-Package Migration](#package-by-package-migration)
  - [config](#config)
  - [middleware](#middleware)
  - [apperror](#apperror)
  - [response / request](#response--request)
  - [database / scan](#database--scan)
  - [cache](#cache)
  - [crypto](#crypto)
  - [validation / sanitize](#validation--sanitize)
  - [logging](#logging)
  - [types (pagination)](#types-pagination)
  - [sorting](#sorting)
  - [health](#health)
  - [retry / circuitbreaker](#retry--circuitbreaker)
- [Breaking Change Policy](#breaking-change-policy)
- [Troubleshooting](#troubleshooting)

---

## Adopting cat-shared-lib

### Prerequisites

| Requirement | Minimum |
|-------------|---------|
| Go          | 1.25+   |
| PostgreSQL  | 14+ (if using `database`/`scan`/`migration`) |

### Step 1 — Add the Dependency

```bash
go get github.com/Saver-Street/cat-shared-lib@latest
```

Import only the packages you need—there is no single top-level import.

### Step 2 — Replace Inline Utilities

Most services have hand-rolled versions of what `cat-shared-lib` provides.
Migrate them one package at a time, starting with the lowest-risk replacements.

**Recommended adoption order** (least to most invasive):

1. `config` — drop-in replacement for `os.Getenv` wrappers
2. `validation` / `sanitize` — replace inline regex checks
3. `crypto` — replace bcrypt/token helpers
4. `response` / `request` — replace JSON encoding helpers
5. `apperror` — replace custom error types
6. `middleware` — replace middleware implementations
7. `database` / `scan` — replace connection and query helpers
8. `logging` / `metrics` / `tracing` — replace observability setup

### Step 3 — Adopt the Middleware Stack

**Before** (typical hand-rolled middleware):

```go
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```

**After:**

```go
import (
	"github.com/Saver-Street/cat-shared-lib/middleware"
)

handler := middleware.Chain(mux,
	middleware.CorrelationID,
	middleware.DetailedLogging(slog.Default()),
	middleware.Auth(jwtSecret),
	middleware.Timeout(5 * time.Second),
	middleware.MaxBody(1 << 20),
)
```

Key differences:
- `middleware.Chain` applies middleware left-to-right (first = outermost)
- All middleware follows the `func(http.Handler) http.Handler` signature
- Correlation IDs propagate through context automatically

### Step 4 — Migrate Error Handling

**Before:**

```go
type APIError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Message }
```

**After:**

```go
import "github.com/Saver-Street/cat-shared-lib/apperror"

// Use constructors instead of struct literals.
err := apperror.NotFound("user", userID)
err := apperror.BadRequest("invalid email")
err := apperror.WrapInternal(dbErr, "query failed")

// In handlers:
var appErr *apperror.Error
if errors.As(err, &appErr) {
	response.JSON(w, appErr.HTTPStatus, appErr)
}
```

Key differences:
- Error codes are standardized strings (e.g., `"NOT_FOUND"`, `"BAD_REQUEST"`)
- HTTP status lives in `HTTPStatus` field (not `Status`)
- `Wrap*` variants preserve the underlying error chain
- `MultiError` collects multiple validation errors

### Step 5 — Adopt Configuration Helpers

**Before:**

```go
port := os.Getenv("PORT")
if port == "" {
	port = "8080"
}
timeout, _ := strconv.Atoi(os.Getenv("TIMEOUT"))
```

**After:**

```go
import "github.com/Saver-Street/cat-shared-lib/config"

port    := config.Get("PORT", "8080")
timeout := config.GetInt("TIMEOUT", 30)
debug   := config.GetBool("DEBUG", false)
dur     := config.GetDuration("INTERVAL", "5s")
```

### Step 6 — Switch to Shared Database Helpers

**Before:**

```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

row := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id=$1", id)
var user User
err := row.Scan(&user.ID, &user.Name)
```

**After:**

```go
import (
	"github.com/Saver-Street/cat-shared-lib/database"
	"github.com/Saver-Street/cat-shared-lib/scan"
)

pool, err := database.Connect(ctx, database.Config{
	URL:          config.Get("DATABASE_URL", ""),
	MaxOpenConns: config.GetInt("DB_MAX_OPEN", 25),
})

user, err := scan.Row[User](ctx, pool, "SELECT id, name FROM users WHERE id=$1", id)
```

Key differences:
- `database.Connect` configures connection pooling out of the box
- `scan.Row` / `scan.Rows` map columns to struct fields generically
- `database.WithTx` handles transaction commit/rollback automatically

### Step 7 — Adopt Observability Packages

**Logging:**

```go
// Before: log.Printf(...)
// After:
import "github.com/Saver-Street/cat-shared-lib/logging"

logger := logging.New(logging.Config{Level: "info", Format: "json"})
logger.Info("user created", "user_id", id)
logger.Error("query failed", logging.Err(err))
```

**Metrics:**

```go
import "github.com/Saver-Street/cat-shared-lib/metrics"

reg := metrics.NewRegistry("myservice")
counter := reg.Counter("requests_total", "Total requests")
counter.Inc()
mux.Handle("GET /metrics", reg.Handler())
```

**Tracing:**

```go
import "github.com/Saver-Street/cat-shared-lib/tracing"

shutdown, err := tracing.Init(ctx, tracing.Config{
	ServiceName: "my-service",
	Endpoint:    config.Get("OTEL_ENDPOINT", ""),
})
defer shutdown(ctx)
```

### Step 8 — Switch to Shared Test Helpers

**Before:**

```go
if got != want {
	t.Errorf("got %v, want %v", got, want)
}
```

**After:**

```go
import "github.com/Saver-Street/cat-shared-lib/testkit"

testkit.AssertEqual(t, want, got)
testkit.AssertTrue(t, condition)
testkit.RequireNoError(t, err)
testkit.AssertStatus(t, http.StatusOK, rec.Code)
testkit.AssertContains(t, body, "expected")
```

---

## Package-by-Package Migration

### config

| Old Pattern | New Pattern |
|---|---|
| `os.Getenv("KEY")` | `config.Get("KEY", "default")` |
| `strconv.Atoi(os.Getenv("N"))` | `config.GetInt("N", 0)` |
| `os.Getenv("FLAG") == "true"` | `config.GetBool("FLAG", false)` |
| `time.ParseDuration(os.Getenv("D"))` | `config.GetDuration("D", "5s")` |

### middleware

| Old Pattern | New Pattern |
|---|---|
| Custom auth middleware | `middleware.Auth(secret)` |
| Custom logging middleware | `middleware.DetailedLogging(logger)` |
| Custom timeout wrapper | `middleware.Timeout(duration)` |
| `http.MaxBytesReader` wrapper | `middleware.MaxBody(bytes)` |
| Custom correlation ID | `middleware.CorrelationID` |
| Manual middleware chaining | `middleware.Chain(handler, mw1, mw2)` |

### apperror

| Old Pattern | New Pattern |
|---|---|
| `&MyError{Status: 404, Msg: "..."}` | `apperror.NotFound("resource", id)` |
| `&MyError{Status: 400, Msg: "..."}` | `apperror.BadRequest("details")` |
| `fmt.Errorf("wrap: %w", err)` | `apperror.WrapInternal(err, "msg")` |
| Collecting errors in a slice | `apperror.MultiError` |

### response / request

| Old Pattern | New Pattern |
|---|---|
| `json.NewEncoder(w).Encode(v)` | `response.JSON(w, status, v)` |
| `w.WriteHeader(204)` | `response.NoContent(w)` |
| `json.NewDecoder(r.Body).Decode(&v)` | `request.JSON(r, &v)` |
| `r.PathValue("id")` + UUID parse | `request.PathUUID(r, "id")` |
| `r.URL.Query().Get("page")` + atoi | `request.QueryInt(r, "page", 1)` |

### database / scan

| Old Pattern | New Pattern |
|---|---|
| `sql.Open(...)` | `database.Connect(ctx, cfg)` |
| `row.Scan(&a, &b, &c)` | `scan.Row[T](ctx, db, query, args)` |
| Manual `BEGIN` / `COMMIT` | `database.WithTx(ctx, db, fn)` |

### cache

| Old Pattern | New Pattern |
|---|---|
| `sync.Map` with manual eviction | `cache.New[K,V](cfg)` |
| Custom LRU implementation | `cache.New[K,V](cfg)` with `MaxSize` and `TTL` |

### crypto

| Old Pattern | New Pattern |
|---|---|
| `bcrypt.GenerateFromPassword(...)` | `crypto.HashPassword(password)` |
| `bcrypt.CompareHashAndPassword(...)` | `crypto.CheckPassword(hash, password)` |
| `crypto/rand` + `hex.Encode` | `crypto.GenerateToken(n)` |

### validation / sanitize

| Old Pattern | New Pattern |
|---|---|
| `regexp.Match(emailRE, input)` | `validation.IsEmail(input)` |
| `uuid.Parse(s) != nil` | `validation.IsUUID(s)` |
| `strings.TrimSpace(s)` | `sanitize.String(s)` |
| `strings.ToLower(strings.TrimSpace(e))` | `sanitize.Email(e)` |
| Custom HTML stripping | `sanitize.HTML(s)` |

### logging

| Old Pattern | New Pattern |
|---|---|
| `log.Printf(...)` | `logger.Info(msg, key, val)` |
| `log.Fatalf(...)` | `logger.Error(msg, logging.Err(err))` |
| Custom JSON log formatter | `logging.New(logging.Config{Format: "json"})` |

### types (pagination)

| Old Pattern | New Pattern |
|---|---|
| Custom cursor pagination struct | `types.NewCursorPage(items, hasMore, cursorFn)` |
| Custom offset pagination struct | `types.NewOffsetPage(items, total, params)` |

### sorting

| Old Pattern | New Pattern |
|---|---|
| Manual `ORDER BY` construction | `sorting.Parse(query, cfg)` then `field.SQL()` |
| Allowlist check for sort fields | `sorting.Config{Allowed: [...]}` |

### health

| Old Pattern | New Pattern |
|---|---|
| `func healthz(w, r) { w.Write("ok") }` | `health.Handler()` |
| Custom dependency checks | `health.NewChecker(health.WithCheck(...))` |

### retry / circuitbreaker

| Old Pattern | New Pattern |
|---|---|
| Manual retry loop with sleep | `retry.Do(ctx, cfg, fn)` |
| Custom circuit breaker | `circuitbreaker.New(cfg)` then `cb.Execute(fn)` |

---

## Breaking Change Policy

`cat-shared-lib` follows these rules for breaking changes:

1. **No breaking changes without a major version bump.** Public API signatures,
   struct fields, and behavior are stable within a major version.

2. **Deprecation before removal.** Deprecated APIs are marked with `// Deprecated:`
   comments and remain for at least one release cycle.

3. **New packages are not breaking.** Adding new packages or new exported
   symbols to existing packages is always backward-compatible.

4. **Default value changes are documented.** If a default changes (e.g.,
   timeout duration), it is called out in the release notes.

---

## Troubleshooting

### "cannot find package"

Ensure your `go.mod` requires the correct version:

```bash
go get github.com/Saver-Street/cat-shared-lib@latest
go mod tidy
```

### "undefined: middleware.CorrelationID"

This was added recently. Update to the latest version:

```bash
go get -u github.com/Saver-Street/cat-shared-lib
```

### Build fails with Go version error

`cat-shared-lib` requires **Go 1.25+**. Check your version:

```bash
go version
```

### Test helpers not found

Import the correct package:

```go
import "github.com/Saver-Street/cat-shared-lib/testkit"      // assertions
import "github.com/Saver-Street/cat-shared-lib/servicetest"   // integration helpers
```

### "apperror.Error has no field Status"

The field is named `HTTPStatus`, not `Status`:

```go
var appErr *apperror.Error
if errors.As(err, &appErr) {
	response.JSON(w, appErr.HTTPStatus, appErr) // ✓ HTTPStatus
}
```

### middleware.Chain order confusion

`middleware.Chain` applies middleware **left-to-right** (first argument is
outermost). Put cross-cutting concerns first:

```go
middleware.Chain(handler,
	middleware.CorrelationID,       // 1st — outermost
	middleware.DetailedLogging(l),  // 2nd
	middleware.Auth(secret),        // 3rd
	middleware.Timeout(5*time.Second), // innermost
)
```

---

*See [USAGE.md](./USAGE.md) for hands-on examples and
[CONTRIBUTING.md](./CONTRIBUTING.md) for development setup.*
