# Usage Guide

This guide walks through common tasks when building Catherine microservices with
`cat-shared-lib`. Each section is self-contained—jump to whatever fits your
current need.

> **Module path:** `github.com/Saver-Street/cat-shared-lib`
>
> **Minimum Go version:** 1.25

---


## Table of Contents

- [Quick Start](#quick-start--minimal-service)
- [Configuration](#configuration)
- [Database](#database)
- [Middleware Stack](#middleware-stack)
- [Request Parsing](#request-parsing)
- [Response Helpers](#response-helpers)
- [Error Handling](#error-handling)
- [Validation](#validation)
- [Full Service Example](#full-service-example)

> For advanced topics, see [Advanced Usage](./ADVANCED-USAGE.md)

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

---

## Related Documentation

- [README](./README.md) — Overview & installation
- [Advanced Usage](./ADVANCED-USAGE.md) — Resilience, health checks, observability, crypto, email, feature flags, migrations, OpenAPI, testing
- [Core Packages](./PACKAGES-GUIDE.md) — Detailed package reference
