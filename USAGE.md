# Usage Guide

This guide walks through common tasks when building Catherine microservices with
`cat-shared-lib`. Each section is self-contained—jump to whatever fits your
current need.

> **Module path:** `github.com/Saver-Street/cat-shared-lib`
>
> **Minimum Go version:** 1.25

---

## Table of Contents

- [Quick Start — Minimal Service](#quick-start--minimal-service)
- [Configuration](#configuration)
- [Database](#database)
- [Middleware Stack](#middleware-stack)
  - [Authentication & Authorization](#authentication--authorization)
  - [Correlation IDs](#correlation-ids)
  - [Request Logging](#request-logging)
  - [Rate Limiting](#rate-limiting)
  - [Timeout & Body Limits](#timeout--body-limits)
  - [CORS](#cors)
- [Request Parsing](#request-parsing)
- [Response Helpers](#response-helpers)
- [Pagination](#pagination)
- [Sorting](#sorting)
- [Error Handling](#error-handling)
- [Validation](#validation)
- [Caching](#caching)
- [Resilience — Retry & Circuit Breaker](#resilience--retry--circuit-breaker)
- [Health Checks](#health-checks)
- [Observability — Logging, Metrics, Tracing](#observability--logging-metrics-tracing)
- [Cryptography](#cryptography)
- [Email](#email)
- [Feature Flags](#feature-flags)
- [Database Migrations](#database-migrations)
- [OpenAPI Spec Generation](#openapi-spec-generation)
- [Integration Testing](#integration-testing)

---

## Quick Start — Minimal Service

A production-ready service in ~40 lines:

```go
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Saver-Street/cat-shared-lib/config"
	"github.com/Saver-Street/cat-shared-lib/health"
	"github.com/Saver-Street/cat-shared-lib/middleware"
	"github.com/Saver-Street/cat-shared-lib/response"
	"github.com/Saver-Street/cat-shared-lib/server"
)

func main() {
	mux := http.NewServeMux()

	// Health endpoint (required for Kubernetes probes).
	mux.HandleFunc("GET /healthz", health.Handler())

	// Application route.
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"message": "Hello, Catherine!"})
	})

	// Middleware stack — order matters (outermost runs first).
	handler := middleware.Chain(mux,
		middleware.CorrelationID,
		middleware.DetailedLogging(slog.Default()),
		middleware.Timeout(config.GetDuration("REQUEST_TIMEOUT", "5s")),
		middleware.MaxBody(1<<20), // 1 MiB
	)

	server.ListenAndServe(server.Config{
		Addr:    config.Get("ADDR", ":8080"),
		Handler: handler,
	})
}
```

---

## Configuration

Load typed values from environment variables with defaults:

```go
import "github.com/Saver-Street/cat-shared-lib/config"

port     := config.Get("PORT", "8080")              // string
debug    := config.GetBool("DEBUG", false)           // bool
workers  := config.GetInt("WORKERS", 4)              // int
timeout  := config.GetDuration("TIMEOUT", "30s")     // time.Duration
```

Every `Get*` function accepts a default value returned when the variable is
unset or empty. No panics, no setup—just call and go.

---

## Database

Open a connection pool and run queries inside transactions:

```go
import "github.com/Saver-Street/cat-shared-lib/database"

pool, err := database.Connect(ctx, database.Config{
	URL:          config.Get("DATABASE_URL", ""),
	MaxOpenConns: config.GetInt("DB_MAX_OPEN", 25),
	MaxIdleConns: config.GetInt("DB_MAX_IDLE", 5),
})
if err != nil {
	log.Fatal(err)
}
defer pool.Close()

// Transactions.
err = database.WithTx(ctx, pool, func(tx database.Tx) error {
	_, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", name)
	return err
})
```

Use `scan.Row` / `scan.Rows` to map results to structs generically:

```go
import "github.com/Saver-Street/cat-shared-lib/scan"

type User struct {
	ID   string
	Name string
}

user, err := scan.Row[User](ctx, pool, "SELECT id, name FROM users WHERE id=$1", id)
users, err := scan.Rows[User](ctx, pool, "SELECT id, name FROM users")
```

---

## Middleware Stack

Build the middleware chain with `middleware.Chain`. Middleware is applied
left-to-right (first argument is outermost):

```go
handler := middleware.Chain(mux,
	middleware.CorrelationID,             // adds X-Correlation-ID
	middleware.DetailedLogging(logger),   // request/response logging
	middleware.Auth(jwtSecret),           // JWT authentication
	middleware.Timeout(5 * time.Second),  // per-request deadline
	middleware.MaxBody(1 << 20),          // 1 MiB body limit
)
```

### Authentication & Authorization

```go
import "github.com/Saver-Street/cat-shared-lib/middleware"

// Require a valid JWT on every request.
authed := middleware.Auth([]byte(config.Get("JWT_SECRET", "")))

// Require specific roles for a route.
adminOnly := middleware.RequireRole("admin")

mux.Handle("DELETE /users/{id}",
	adminOnly(http.HandlerFunc(deleteUserHandler)),
)
```

Access the authenticated user inside a handler:

```go
import "github.com/Saver-Street/cat-shared-lib/identity"

func handler(w http.ResponseWriter, r *http.Request) {
	userID, ok := identity.UserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	// userID is the UUID string from the JWT claims.
}
```

### Correlation IDs

Every request automatically gets an `X-Correlation-ID` header. Downstream
calls propagate it via context:

```go
import "github.com/Saver-Street/cat-shared-lib/middleware"

// In the middleware chain:
middleware.CorrelationID

// In a handler — retrieve the ID:
corrID := middleware.GetCorrelationID(r.Context())
```

### Request Logging

Log every request with method, path, status, duration, and body size:

```go
middleware.DetailedLogging(slog.Default())
```

Logs are structured JSON via `log/slog`, e.g.:

```json
{"level":"INFO","msg":"request completed","method":"GET","path":"/users","status":200,"duration":"1.23ms","bytes":512}
```

### Rate Limiting

```go
import "github.com/Saver-Street/cat-shared-lib/ratelimit"

limiter := ratelimit.New(ratelimit.Config{
	Rate:  100,           // tokens per interval
	Burst: 20,            // max burst
	Per:   time.Minute,   // refill interval
})

mux.Handle("POST /api/submit",
	limiter.Middleware()(http.HandlerFunc(submitHandler)),
)
```

### Timeout & Body Limits

```go
// Per-request timeout (context deadline + 503 on expiry).
middleware.Timeout(5 * time.Second)

// Cap request body size (413 if exceeded).
middleware.MaxBody(1 << 20) // 1 MiB
```

### CORS

```go
import "github.com/Saver-Street/cat-shared-lib/cors"

corsMiddleware := cors.New(cors.Config{
	AllowedOrigins:   []string{"https://app.example.com"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	AllowedHeaders:   []string{"Authorization", "Content-Type"},
	AllowCredentials: true,
	MaxAge:           3600,
})
```

---

## Request Parsing

Parse path parameters, query strings, and JSON bodies:

```go
import "github.com/Saver-Street/cat-shared-lib/request"

func handler(w http.ResponseWriter, r *http.Request) {
	// UUID path parameter.
	id, err := request.PathUUID(r, "id")

	// Query string values.
	page  := request.QueryInt(r, "page", 1)
	limit := request.QueryInt(r, "limit", 20)
	q     := request.QueryString(r, "q", "")

	// Parse & validate JSON body.
	var body CreateUserRequest
	if err := request.JSON(r, &body); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
}
```

### Parsing UUID Lists

```go
// Parse comma-separated UUIDs from a query parameter.
ids, err := request.ParseIDList(r, "ids")
// e.g., ?ids=550e8400-e29b-41d4-a716-446655440000,6ba7b810-9dad-11d1-80b4-00c04fd430c8
```

---

## Response Helpers

```go
import "github.com/Saver-Street/cat-shared-lib/response"

// JSON response.
response.JSON(w, http.StatusOK, user)

// Error response ({"error":"message"}).
response.Error(w, http.StatusNotFound, "user not found")

// No content.
response.NoContent(w)

// Created with Location header.
response.Created(w, "/users/"+user.ID, user)
```

---

## Pagination

### Cursor-Based (Recommended)

```go
import "github.com/Saver-Street/cat-shared-lib/types"

// Build a cursor page from a slice of items.
page := types.NewCursorPage(users, hasMore, func(u User) string {
	return u.ID // cursor extractor
})
// page.Items, page.NextCursor, page.HasMore

response.JSON(w, http.StatusOK, page)
```

### Offset-Based

```go
params := types.OffsetParams{Page: 1, PageSize: 20}
page := types.NewOffsetPage(users, totalCount, params)
// page.Items, page.TotalCount, page.Page, page.PageSize, page.TotalPages

response.JSON(w, http.StatusOK, page)
```

---

## Sorting

```go
import "github.com/Saver-Street/cat-shared-lib/sorting"

cfg := sorting.Config{
	Allowed:          []string{"name", "created_at", "updated_at"},
	DefaultField:     "created_at",
	DefaultDirection: sorting.Desc,
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	field, err := sorting.Parse(r.URL.Query(), cfg)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	orderClause := field.SQL() // e.g., "created_at DESC"
	// Use in your SQL query.
}
```

---

## Error Handling

### Structured Errors

```go
import "github.com/Saver-Street/cat-shared-lib/apperror"

// Pre-defined error constructors for common HTTP errors.
err := apperror.NotFound("user", userID)
err := apperror.BadRequest("invalid email format")
err := apperror.Unauthorized("token expired")
err := apperror.Forbidden("insufficient permissions")
err := apperror.Conflict("email already exists")
err := apperror.InternalError("database connection failed")

// Custom error with code.
err := apperror.New(http.StatusUnprocessableEntity, "VALIDATION_FAILED", "name is required")

// Wrap underlying errors.
err := apperror.WrapNotFound(dbErr, "user", userID)
err := apperror.WrapInternal(dbErr, "failed to query users")
```

### Error Response Pattern

```go
func handler(w http.ResponseWriter, r *http.Request) {
	user, err := repo.GetUser(ctx, id)
	if err != nil {
		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			response.JSON(w, appErr.HTTPStatus, appErr)
		} else {
			response.Error(w, 500, "internal error")
		}
		return
	}
	response.JSON(w, 200, user)
}
```

### Collecting Multiple Errors

```go
var multi apperror.MultiError
if name == "" {
	multi.Add(apperror.BadRequest("name is required"))
}
if email == "" {
	multi.Add(apperror.BadRequest("email is required"))
}
if err := multi.Err(); err != nil {
	response.JSON(w, http.StatusBadRequest, err)
	return
}
```

---

## Validation

```go
import "github.com/Saver-Street/cat-shared-lib/validation"

// Built-in validators.
ok := validation.IsEmail("user@example.com")
ok := validation.IsUUID("550e8400-e29b-41d4-a716-446655440000")
ok := validation.IsURL("https://example.com")
ok := validation.InRange(age, 1, 150)
ok := validation.MaxLength(name, 255)
```

### Input Sanitization

```go
import "github.com/Saver-Street/cat-shared-lib/sanitize"

clean := sanitize.String(userInput)       // trim + normalize whitespace
clean := sanitize.Email(rawEmail)         // lowercase + trim
clean := sanitize.HTML(untrusted)         // strip dangerous tags
```

### PII Redaction

```go
import "github.com/Saver-Street/cat-shared-lib/security"

safe := security.RedactPII(logMessage)    // mask emails, SSNs, etc.
```

---

## Caching

```go
import "github.com/Saver-Street/cat-shared-lib/cache"

c := cache.New[string, User](cache.Config{
	MaxSize: 1000,
	TTL:     5 * time.Minute,
})

c.Set("user:123", user)

if u, ok := c.Get("user:123"); ok {
	// cache hit
}

c.Delete("user:123")
```

The cache is thread-safe, uses LRU eviction, and respects TTL expiry.

---

## Resilience — Retry & Circuit Breaker

### Retry with Exponential Backoff

```go
import "github.com/Saver-Street/cat-shared-lib/retry"

result, err := retry.Do(ctx, retry.Config{
	MaxAttempts: 3,
	InitialWait: 100 * time.Millisecond,
	MaxWait:     2 * time.Second,
	Multiplier:  2.0,
}, func(ctx context.Context) (string, error) {
	return callExternalService(ctx)
})
```

### Circuit Breaker

```go
import "github.com/Saver-Street/cat-shared-lib/circuitbreaker"

cb := circuitbreaker.New(circuitbreaker.Config{
	MaxFailures:  5,
	ResetTimeout: 30 * time.Second,
})

err := cb.Execute(func() error {
	return callDownstream()
})
// Returns circuitbreaker.ErrOpen when the circuit is open.
```

---

## Health Checks

### Basic

```go
import "github.com/Saver-Street/cat-shared-lib/health"

mux.HandleFunc("GET /healthz", health.Handler())
```

### With Dependency Checks

```go
checker := health.NewChecker(
	health.WithCheck("postgres", func(ctx context.Context) error {
		return pool.Ping(ctx)
	}),
	health.WithCheck("redis", func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	}),
)

mux.HandleFunc("GET /healthz", checker.Handler())
```

### Service Discovery

```go
import "github.com/Saver-Street/cat-shared-lib/discovery"

registry := discovery.NewRegistry()
registry.Register("user-service", "http://user-svc:8080")
registry.Register("order-service", "http://order-svc:8080")

addr, err := registry.Resolve("user-service")
```

---

## Observability — Logging, Metrics, Tracing

### Structured Logging

```go
import "github.com/Saver-Street/cat-shared-lib/logging"

logger := logging.New(logging.Config{
	Level:  "info",
	Format: "json",
})

logger.Info("user created", "user_id", id, "email", email)
logger.Error("failed to send email", logging.Err(err))
```

### Prometheus Metrics

```go
import "github.com/Saver-Street/cat-shared-lib/metrics"

reg := metrics.NewRegistry("myservice")
counter := reg.Counter("requests_total", "Total HTTP requests")
histogram := reg.Histogram("request_duration_seconds", "Request latency")

counter.Inc()
histogram.Observe(elapsed.Seconds())

// Expose /metrics endpoint.
mux.Handle("GET /metrics", reg.Handler())
```

### Distributed Tracing

```go
import "github.com/Saver-Street/cat-shared-lib/tracing"

shutdown, err := tracing.Init(ctx, tracing.Config{
	ServiceName: "user-service",
	Endpoint:    config.Get("OTEL_ENDPOINT", "localhost:4317"),
})
if err != nil {
	log.Fatal(err)
}
defer shutdown(ctx)

// Spans are created automatically by the tracing middleware.
// For manual spans:
ctx, span := tracing.Start(ctx, "fetchUser")
defer span.End()
```

---

## Cryptography

```go
import "github.com/Saver-Street/cat-shared-lib/crypto"

// Password hashing (bcrypt).
hash, err := crypto.HashPassword("s3cret")
ok := crypto.CheckPassword(hash, "s3cret") // true

// Secure random tokens.
token, err := crypto.GenerateToken(32)     // 32-byte hex-encoded token
```

---

## Email

```go
import "github.com/Saver-Street/cat-shared-lib/email"

mailer := email.NewMailer(email.Config{
	Host:     config.Get("SMTP_HOST", "localhost"),
	Port:     config.GetInt("SMTP_PORT", 587),
	Username: config.Get("SMTP_USER", ""),
	Password: config.Get("SMTP_PASS", ""),
})

err := mailer.Send(ctx, email.Message{
	To:      []string{"user@example.com"},
	Subject: "Welcome!",
	HTML:    "<h1>Hello</h1><p>Welcome to Catherine.</p>",
	Text:    "Hello\n\nWelcome to Catherine.",
})
```

---

## Feature Flags

### Environment-Based (Simple)

```go
import "github.com/Saver-Street/cat-shared-lib/featureflags"

if featureflags.IsEnabled("NEW_CHECKOUT_FLOW") {
	// new code path
}
```

### Database-Backed (Dynamic)

```go
import "github.com/Saver-Street/cat-shared-lib/flags"

flagStore := flags.New(pool) // pool is your database connection

enabled, err := flagStore.IsEnabled(ctx, "dark_mode")
```

---

## Database Migrations

```go
import "github.com/Saver-Street/cat-shared-lib/migration"

runner := migration.NewRunner(pool, migration.Config{
	Dir: "migrations",
})

if err := runner.Up(ctx); err != nil {
	log.Fatal(err)
}
```

Migration files follow the naming pattern `001_create_users.up.sql` /
`001_create_users.down.sql`.

---

## OpenAPI Spec Generation

Build OpenAPI 3.0 specs programmatically:

```go
import "github.com/Saver-Street/cat-shared-lib/openapi"

spec := openapi.NewSpec("User Service", "1.0.0").
	Description("User management API").
	Server("https://api.example.com").
	Path("/users", openapi.GET, openapi.Operation{
		Summary: "List users",
		Tags:    []string{"users"},
	}).
	Path("/users/{id}", openapi.GET, openapi.Operation{
		Summary:    "Get user by ID",
		Tags:       []string{"users"},
		Parameters: []openapi.Parameter{{Name: "id", In: "path", Required: true}},
	})

yamlBytes, err := spec.YAML()
jsonBytes, err := spec.JSON()
```

---

## Integration Testing

The `servicetest` and `testkit` packages provide helpers for writing clean,
concise tests.

### Assertions

```go
import "github.com/Saver-Street/cat-shared-lib/testkit"

func TestCreateUser(t *testing.T) {
	user := createUser(t, "Alice")
	testkit.AssertEqual(t, "Alice", user.Name)
	testkit.AssertTrue(t, user.ID != "")
	testkit.RequireNoError(t, err)
}
```

### HTTP Testing

```go
func TestListUsers(t *testing.T) {
	rec := testkit.DoRequest(t, handler, "GET", "/users", nil)
	testkit.AssertStatus(t, http.StatusOK, rec.Code)
	testkit.AssertContains(t, rec.Body.String(), "Alice")
}
```

### Integration Test Setup

```go
import "github.com/Saver-Street/cat-shared-lib/servicetest"

func TestIntegration(t *testing.T) {
	env := servicetest.NewEnv(t)

	// env provides a test database, config, and cleanup.
	pool := env.DB()
	// Run your tests against the real database.
}
```

---

## Full Service Example

Putting it all together—a complete microservice skeleton:

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Saver-Street/cat-shared-lib/config"
	"github.com/Saver-Street/cat-shared-lib/cors"
	"github.com/Saver-Street/cat-shared-lib/database"
	"github.com/Saver-Street/cat-shared-lib/health"
	"github.com/Saver-Street/cat-shared-lib/logging"
	"github.com/Saver-Street/cat-shared-lib/middleware"
	"github.com/Saver-Street/cat-shared-lib/response"
	"github.com/Saver-Street/cat-shared-lib/server"
	"github.com/Saver-Street/cat-shared-lib/tracing"
)

func main() {
	ctx := context.Background()

	// Logging.
	logger := logging.New(logging.Config{
		Level:  config.Get("LOG_LEVEL", "info"),
		Format: "json",
	})

	// Tracing.
	shutdownTracer, _ := tracing.Init(ctx, tracing.Config{
		ServiceName: "my-service",
		Endpoint:    config.Get("OTEL_ENDPOINT", ""),
	})

	// Database.
	pool, err := database.Connect(ctx, database.Config{
		URL: config.Get("DATABASE_URL", ""),
	})
	if err != nil {
		logger.Error("database connection failed", logging.Err(err))
		return
	}
	defer pool.Close()

	// Routes.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Handler())
	mux.HandleFunc("GET /api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"hello": "world"})
	})

	// Middleware.
	handler := middleware.Chain(mux,
		cors.New(cors.Config{AllowedOrigins: []string{"*"}}),
		middleware.CorrelationID,
		middleware.DetailedLogging(logger),
		middleware.Auth([]byte(config.Get("JWT_SECRET", ""))),
		middleware.Timeout(config.GetDuration("REQUEST_TIMEOUT", "10s")),
		middleware.MaxBody(1<<20),
	)

	// Start server.
	server.ListenAndServe(server.Config{
		Addr:    config.Get("ADDR", ":8080"),
		Handler: handler,
	}, func() {
		shutdownTracer(ctx)
	})
}
```

---

*For API reference, run `go doc github.com/Saver-Street/cat-shared-lib/<package>` or browse
the [package index](https://pkg.go.dev/github.com/Saver-Street/cat-shared-lib).*
