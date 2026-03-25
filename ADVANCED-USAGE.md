# Advanced Usage Guide

> See also: [Usage Guide](./USAGE.md) | [README](./README.md) | [Core Packages](./PACKAGES-GUIDE.md)

## Table of Contents

- [Pagination](#pagination)
- [Sorting](#sorting)
- [Resilience — Retry & Circuit Breaker](#resilience--retry--circuit-breaker)
- [Health Checks](#health-checks)
- [Observability](#observability--logging-metrics-tracing)
- [Cryptography](#cryptography)
- [Email](#email)
- [Feature Flags](#feature-flags)
- [Database Migrations](#database-migrations)
- [OpenAPI Spec Generation](#openapi-spec-generation)
- [Caching](#caching)
- [Integration Testing](#integration-testing)


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



## Related Documentation

- [Usage Guide](./USAGE.md) — Quick start, common patterns
- [README](./README.md) — Overview & installation
- [Core Packages](./PACKAGES-GUIDE.md) — Detailed package reference
- [Migration Guide](./MIGRATION.md) — Adoption steps
