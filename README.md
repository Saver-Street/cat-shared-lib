# cat-shared-lib

[![CI](https://github.com/Saver-Street/cat-shared-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/Saver-Street/cat-shared-lib/actions/workflows/ci.yml)

Shared Go library for Catherine (Auto-Apps) microservices.

**Wave 0 — required by all service extractions.**

---

## Table of Contents

- [Installation](#installation)
- [Packages](#packages)
- [Usage Examples](#usage-examples)
  - [middleware — JWT Auth, Authorization, Request Handling](#middleware--jwt-auth-authorization-request-handling)
  - [config — Environment Variables](#config--environment-variables)
  - [database — PostgreSQL Pool & Transactions](#database--postgresql-pool--transactions)
  - [validation — Input Validation](#validation--input-validation)
  - [cache — In-Memory LRU Cache with TTL](#cache--in-memory-lru-cache-with-ttl)
  - [retry — Exponential Backoff](#retry--exponential-backoff)
  - [crypto — Password Hashing & Token Generation](#crypto--password-hashing--token-generation)
  - [email — SMTP Mailer with Templates](#email--smtp-mailer-with-templates)
  - [tracing — OpenTelemetry Distributed Tracing](#tracing--opentelemetry-distributed-tracing)
  - [migration — Database Migration Runner](#migration--database-migration-runner)
  - [response — JSON Response Helpers](#response--json-response-helpers)
  - [request — HTTP Request Parsing](#request--http-request-parsing)
  - [health — Health Check Handlers](#health--health-check-handlers)
  - [httpclient — Resilient HTTP Client](#httpclient--resilient-http-client)
  - [apperror — Structured Errors](#apperror--structured-errors)
  - [Additional Packages](#additional-packages)
  - [contracts — Service Interface Contracts](#contracts--service-interface-contracts)
  - [servicetest — Integration Test Helpers](#servicetest--integration-test-helpers)
- [Design Notes](#design-notes)
- [Testing](#testing)

---

## Installation

```sh
go get github.com/Saver-Street/cat-shared-lib@v1.0.0
```

```go
// go.mod
require github.com/Saver-Street/cat-shared-lib v1.0.0
```

---

## Packages

| Package | Description | Coverage |
|---|---|---|
| `middleware` | JWT auth (HS256), request ID, logging, recovery, rate limiting, ETag, caching | 99.7% |
| `config` | Env var parsing with defaults, validation, byte sizes, URLs, ports | 100% |
| `database` | PostgreSQL connection pool, transaction helpers | 96.6% |
| `validation` | Email, UUID, phone, URL, slug, alphanumeric, hex, numeric, range validators | 100% |
| `cache` | Generic in-memory LRU cache with per-entry TTL | 98% |
| `retry` | Exponential backoff with jitter and context cancellation | 100% |
| `crypto` | bcrypt password hashing, secure token generation, HMAC-SHA256, SHA-256 | 100% |
| `email` | SMTP mailer with HTML/text template support | 92.6% |
| `tracing` | OpenTelemetry distributed tracing setup and helpers | 98.5% |
| `migration` | Database migration runner with rollback support | 100% |
| `response` | JSON response helpers, pagination, SSE, file downloads, streaming | 100% |
| `request` | HTTP request parsing, URL param extraction, JSON body decoding | 100% |
| `health` | Standardized health check handlers with concurrent checkers | 99.2% |
| `httpclient` | Resilient HTTP client with retries and circuit breaker | 100% |
| `apperror` | Standardized error types with HTTP status codes | 100% |
| `circuitbreaker` | Circuit breaker pattern for resilient service calls | 100% |
| `ratelimit` | Per-key sliding-window + token-bucket rate limiter | 100% |
| `cors` | CORS middleware with configurable origins | 100% |
| `discovery` | Service registry and instance-based health checking | 100% |
| `entitlements` | Subscription tier limits + DirectDB queries | 100% |
| `featureflags` | Environment-variable-based feature flags | 100% |
| `flags` | Feature flag DirectDB reads (boolean/plain-text) | 100% |
| `identity` | Candidate resolution + context getters | 100% |
| `metrics` | Prometheus metrics helpers | 100% |
| `openapi` | OpenAPI/Swagger spec serving | 100% |
| `sanitize` | Filename sanitization, HTML escaping, whitespace normalization, Deref | 100% |
| `scan` | Generic database row scanning (Rows, Row, First) | 100% |
| `security` | Suspicious input detection, PII redaction, URL credential scrubbing | 100% |
| `server` | HTTP server with graceful shutdown (SIGINT/SIGTERM) | 100% |
| `shutdown` | OS signal-based graceful shutdown coordinator | 100% |
| `testkit` | Mock server, call recorder, and other test helpers | 100% |
| `types` | Shared domain types (User, CandidateProfile, Pagination) | 100% |
| `contracts` | Shared service interfaces (Service, Handler, HealthCheck, StandardError) | 100% |
| `servicetest` | Integration test helpers: HTTP test server, mock Querier, fixture loader | 100% |

---

## Usage Examples

### middleware — JWT Auth, Authorization, Request Handling

The `middleware` package provides cross-cutting HTTP middleware: JWT validation,
authorization checks, request ID injection, structured logging, panic recovery,
rate limiting, and brute-force protection.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/middleware"
    "os"
    "time"
)

// --- JWT Auth middleware ---
authMW := middleware.JWTAuth(middleware.JWTAuthConfig{
    Secret:    []byte(os.Getenv("JWT_SECRET")),
    Issuer:    "cat-service",
    SkipPaths: []string{"/health", "/ready"},
})

mux := http.NewServeMux()
mux.Handle("/api/", authMW(apiHandler))

// Read authenticated user info inside any handler:
userID   := middleware.GetUserID(r)
email    := middleware.GetUserEmail(r)
role     := middleware.GetUserRole(r)
tier     := middleware.GetSubscriptionTier(r)

// Issue a signed JWT:
token, err := middleware.SignHS256(middleware.JWTClaims{
    Subject:   userID,
    Email:     email,
    Role:      "admin",
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    Issuer:    "cat-service",
}, []byte(os.Getenv("JWT_SECRET")))

// --- Authorization middleware ---
mux.Handle("/api/",    middleware.RequireAuth(handler))
mux.Handle("/admin/",  middleware.RequireAdmin(handler))
mux.Handle("/editor/", middleware.RequireRole("editor")(handler))
mux.Handle("/premium/",middleware.RequireSubscriptionTier("pro")(handler))

// --- Stack middleware (outermost first) ---
handler := middleware.Chain(
    middleware.Recovery,
    middleware.RequestID,
    middleware.SecureHeaders,       // X-Content-Type-Options, X-Frame-Options
    middleware.HSTS(365*24*time.Hour, true), // Strict-Transport-Security
    middleware.NoCache,             // Cache-Control, Pragma, Expires
    middleware.ETag,                // Automatic ETag + 304 Not Modified
    middleware.MaxBody(1<<20),      // 1 MiB request body limit
    middleware.Timeout(30*time.Second),
    middleware.Logging(logger),
    middleware.AllowMethods("GET", "POST", "PUT", "DELETE"),
    middleware.ContentType("application/json"), // enforce JSON for POST/PUT/PATCH
    middleware.Compress,            // gzip response compression
    middleware.RealIP,              // normalize X-Real-IP from proxy headers
)(apiHandler)

// --- Per-route caching ---
mux.Handle("/static/", middleware.CacheControl(time.Hour, true)(staticHandler))
```

---

### config — Environment Variables

The `config` package reads configuration from environment variables with support
for defaults, type coercion, required checks, and bulk validation.

```go
import "github.com/Saver-Street/cat-shared-lib/config"

// Validate required vars at startup — fail fast
if err := config.Validate("DATABASE_URL", "JWT_SECRET"); err != nil {
    log.Fatal(err) // "config: missing required env vars: JWT_SECRET"
}

port    := config.Int("PORT", 8080)
dbURL   := config.MustString("DATABASE_URL")    // panics if unset
debug   := config.Bool("DEBUG", false)
timeout := config.Duration("TIMEOUT", 30*time.Second)
origins := config.StringSlice("CORS_ORIGINS", []string{"http://localhost:3000"})

// Additional typed helpers
apiURL, err  := config.URL("API_URL", "https://api.example.com")
listenAddr, err := config.Addr("LISTEN_ADDR", ":8080")
maxUpload, err := config.Bytes("MAX_UPLOAD", 10*1024*1024) // supports "64MB"
tcpPort, err := config.Port("TCP_PORT", 3000)

// Panic-free required string (returns error instead)
secret, err := config.StringRequired("JWT_SECRET")
allowed, err := config.StringSliceRequired("ALLOWED_ORIGINS")

// Enum-style constraints
logLevel, err := config.Enum("LOG_LEVEL", "info", []string{"debug", "info", "warn", "error"})
mode := config.MustEnum("APP_MODE", []string{"development", "staging", "production"})
apiBase := config.MustURL("API_BASE_URL")
```

---

### database — PostgreSQL Pool & Transactions

The `database` package wraps `pgx/v5` to provide connection pooling and
transaction helpers. DB-querying functions accept a `Querier` interface so they
work identically with a pool, connection, or open transaction.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/database"
    "github.com/jackc/pgx/v5"
)

// Open a connection pool
pool, err := database.NewPool(ctx, database.PoolConfig{
    DSN:             config.MustString("DATABASE_URL"),
    MaxConns:        20,
    MinConns:        5,
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: 30 * time.Minute,
})
defer pool.Close()

// Run a transaction — automatically rolled back on error or panic
err = database.WithTx(ctx, pool, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx,
        "INSERT INTO users (id, email) VALUES ($1, $2)", id, email)
    if err != nil {
        return err // triggers rollback
    }
    _, err = tx.Exec(ctx,
        "INSERT INTO profiles (user_id) VALUES ($1)", id)
    return err // nil → committed
})
```

---

### validation — Input Validation

The `validation` package provides field-level validators that return structured
`ValidationError` values with the field name and a human-readable message.

```go
import "github.com/Saver-Street/cat-shared-lib/validation"

errs := validation.Collect(
    validation.Required("name",     req.Name),
    validation.Email("email",       req.Email),
    validation.UUID("userId",       req.UserID),
    validation.Phone("phone",       req.Phone),
    validation.URL("website",       req.Website),
    validation.MinLength("password",req.Password, 8),
    validation.MaxLength("bio",     req.Bio, 500),
    validation.ExactLength("pin",   req.Pin, 6),
    validation.OneOf("role",        req.Role, []string{"admin","user","guest"}),
    validation.Slug("handle",       req.Handle),
    validation.NoWhitespace("apiKey", req.APIKey),
    validation.Alphanumeric("code", req.Code),
    validation.Numeric("zip",       req.Zip),
    validation.Hex("color",         req.Color),
    validation.Between("age",       req.Age, 18, 120),
    validation.Lowercase("slug",    req.Slug),
    validation.Uppercase("country", req.Country),
    validation.StartsWith("path",   req.Path, "/api/"),
    validation.EndsWith("file",     req.File, ".pdf"),
)
if len(errs) > 0 {
    // Each error is a *validation.ValidationError with .Field and .Message
    response.BadRequest(w, errs[0].Error())
    return
}

// Validate all elements in a slice
err := validation.EachString("tags", req.Tags, validation.Alphanumeric)
```

---

### cache — In-Memory LRU Cache with TTL

The `cache` package provides a generic, thread-safe LRU cache with per-entry TTL
and a background cleanup goroutine to evict expired entries.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/cache"
    "time"
)

// Create a cache of up to 1 000 string→User entries
c := cache.New[string, User](cache.Config{
    MaxEntries:      1000,
    DefaultTTL:      5 * time.Minute,
    CleanupInterval: time.Minute,
})
defer c.Stop() // stops the background cleanup goroutine

// Store
c.Set("user:42", user)
c.SetWithTTL("session:abc", session, 30*time.Minute)

// Retrieve
u, ok := c.Get("user:42")
if !ok {
    // cache miss — load from DB
}

// Invalidate
c.Delete("user:42")
c.Clear() // remove all entries

fmt.Println(c.Len())      // current entry count
fmt.Println(c.Contains("user:42")) // check existence without LRU promotion

// Lazy-load pattern
user := c.GetOrSet("user:42", func() User {
    return fetchUserFromDB("42")
})
```

---

### retry — Exponential Backoff

The `retry` package retries operations with exponential backoff and optional
full jitter. It respects `context.Context` cancellation and supports custom
retry predicates.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/retry"
    "time"
)

err := retry.Do(ctx, retry.Config{
    MaxAttempts:   5,
    InitialDelay:  100 * time.Millisecond,
    MaxDelay:      10 * time.Second,
    Multiplier:    2.0,
    JitterFraction: 0.25,
    // Only retry on network errors, not on business logic errors:
    RetryIf: func(err error) bool {
        return !errors.Is(err, ErrNotFound)
    },
}, func(ctx context.Context) error {
    return callExternalService(ctx)
})
if err != nil {
    // all attempts exhausted or context cancelled
}

// Calculate delay for a given attempt (for logging)
d := retry.Delay(retry.Config{InitialDelay: 100*time.Millisecond}, 2) // 400ms
```

---

### crypto — Password Hashing & Token Generation

The `crypto` package provides bcrypt password hashing, secure random token
generation, and HMAC-SHA256 signing with constant-time comparison.

```go
import "github.com/Saver-Street/cat-shared-lib/crypto"

// --- Password hashing ---
hash, err := crypto.HashPassword("super-secret")  // default bcrypt cost
if err != nil { ... }

err = crypto.CheckPassword("super-secret", hash)
if errors.Is(err, crypto.ErrInvalidToken) {
    // wrong password
}

// Rehash if stored with an old/weak cost
if crypto.NeedsRehash(hash) {
    newHash, _ := crypto.HashPassword(plaintextPassword)
    // update DB record
}

// --- Secure random tokens ---
token, err := crypto.GenerateToken(32)     // 32 random bytes → base64 string
hexToken, err := crypto.GenerateHexToken(16) // 16 random bytes → hex string

// --- HMAC-SHA256 signing ---
mac := crypto.HMACSHA256([]byte("my-secret-key"), []byte("payload"))
ok  := crypto.VerifyHMACSHA256([]byte("my-secret-key"), []byte("payload"), mac)

// --- Utility ---
code := crypto.RandomString(6)   // cryptographically random alphanumeric
hash := crypto.HashSHA256(data)  // hex-encoded SHA-256
```

---

### email — SMTP Mailer with Templates

The `email` package sends emails over SMTP with full MIME multipart support
(HTML + plain-text alternatives), quoted-printable encoding, and Go template
rendering for both HTML and text bodies.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/email"
    "context"
)

// Create a mailer
mailer := email.NewMailer(email.Config{
    Host:     "smtp.sendgrid.net",
    Port:     587,
    Username: os.Getenv("SMTP_USER"),
    Password: os.Getenv("SMTP_PASS"),
    From:     "no-reply@example.com",
    Timeout:  30 * time.Second,
})

// Send a plain email
err := mailer.Send(ctx, email.Message{
    To:      []string{"user@example.com"},
    CC:      []string{"manager@example.com"},
    Subject: "Welcome to Catherine",
    HTML:    "<h1>Hello!</h1><p>Your account is ready.</p>",
    Text:    "Hello!\n\nYour account is ready.",
    Headers: map[string]string{"X-Priority": "1"},
})

// Send with Go templates
htmlTmpl, _ := email.ParseHTMLString("welcome",
    `<h1>Hello, {{.Name}}!</h1><p>Your code is {{.Code}}.</p>`)
textTmpl, _ := email.ParseTextString("welcome",
    `Hello, {{.Name}}! Your code is {{.Code}}.`)

data := map[string]string{"Name": "Jordan", "Code": "ABC-123"}

htmlBody, _ := email.RenderHTML(htmlTmpl, "welcome", data)
textBody, _ := email.RenderText(textTmpl, "welcome", data)

err = mailer.Send(ctx, email.Message{
    To:      []string{"jordan@example.com"},
    Subject: "Your verification code",
    HTML:    htmlBody,
    Text:    textBody,
})
```

---

### tracing — OpenTelemetry Distributed Tracing

The `tracing` package wraps the OTel SDK to make trace setup, span creation,
HTTP propagation, and context management simple and consistent across services.

```go
import "github.com/Saver-Street/cat-shared-lib/tracing"

// --- Initialise once at service startup ---
tp, err := tracing.NewProvider(ctx, tracing.Config{
    ServiceName:    "billing-service",
    ServiceVersion: "1.2.0",
    Environment:    "production",
    Exporter:       tracing.ExporterStdout, // or ExporterNoop
    SampleRate:     1.0,
})
if err != nil { log.Fatal(err) }
defer tp.Shutdown(ctx)

// --- Create and use spans ---
tracer := tracing.Tracer("billing-service")

ctx, span := tracing.Start(ctx, "process-payment",
    trace.WithAttributes(attribute.String("payment.id", paymentID)),
)
defer span.End()

if err := processPayment(ctx, payment); err != nil {
    tracing.RecordError(span, err)
    return err
}

// --- HTTP middleware for automatic trace propagation ---
mux.Handle("/", tracing.Middleware("billing-service")(handler))

// --- Propagate to outbound HTTP calls ---
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
tracing.InjectHTTP(ctx, req.Header) // injects W3C traceparent header

// --- Access IDs for logging correlation ---
traceID := tracing.TraceID(ctx)
spanID  := tracing.SpanID(ctx)
```

---

### migration — Database Migration Runner

The `migration` package runs ordered SQL migrations tracked in a
`schema_migrations` table. It supports forward migrations, optional rollbacks,
and transaction-safe execution.

```go
import "github.com/Saver-Street/cat-shared-lib/migration"

// Create a migrator (uses "schema_migrations" table by default)
m := migration.New(pool,
    migration.Migration{
        ID:   "001_create_users",
        Up:   "CREATE TABLE users (id UUID PRIMARY KEY, email TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ DEFAULT now())",
        Down: "DROP TABLE users",
    },
    migration.Migration{
        ID:   "002_add_role",
        Up:   "ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'",
        Down: "ALTER TABLE users DROP COLUMN role",
    },
    migration.Migration{
        ID:   "003_create_sessions",
        Up:   `CREATE TABLE sessions (id UUID PRIMARY KEY, user_id UUID REFERENCES users(id), expires_at TIMESTAMPTZ)`,
        Down: "DROP TABLE sessions",
    },
)

// Initialise the tracking table and run all pending migrations
if err := m.Init(ctx); err != nil { log.Fatal(err) }
if err := m.Migrate(ctx); err != nil { log.Fatal(err) }

// Check status
records, _ := m.Status(ctx)
for _, r := range records {
    fmt.Printf("%s: applied=%v at=%v\n", r.ID, r.Applied, r.AppliedAt)
}

// Rollback the last applied migration
if err := m.Rollback(ctx); err != nil { log.Fatal(err) }

// Use a custom tracking table
m = migration.NewWithTable(pool, "my_schema.migrations", migrations...)
```

---

### response — JSON Response Helpers

The `response` package provides uniform JSON response construction with standard
HTTP status codes and pagination support.

```go
import "github.com/Saver-Street/cat-shared-lib/response"

response.OK(w, user)                               // 200 JSON
response.Created(w, newRecord)                      // 201 JSON
response.NoContent(w)                               // 204
response.BadRequest(w, "invalid email")             // 400 JSON error
response.Unauthorized(w, "token expired")           // 401 JSON error
response.Forbidden(w, "insufficient permissions")   // 403 JSON error
response.NotFound(w, "user not found")              // 404 JSON error
response.InternalError(w, "query failed", err)      // 500 (logs error, returns generic message)

// Paginated list with X-Total-Count / X-Limit / X-Offset headers
response.PaginatedWithHeaders(w, users, total, page, limit)

// Decode request body or send 400 automatically
var req CreateUserRequest
if !response.DecodeOrFail(w, r, &req) {
    return // 400 already written
}

// XML response
response.XML(w, http.StatusOK, xmlData)
```

---

### request — HTTP Request Parsing

```go
import "github.com/Saver-Street/cat-shared-lib/request"

// Pagination from query params (default limit, max limit)
p := request.ParsePagination(r.URL.Query(), 20, 100)
// p.Page, p.Limit, p.Offset

// Required URL path parameter (works with any router)
id, err := request.RequireURLParamInt(r, "id", chi.URLParam)

// Optional query param with default
status := request.OptionalQueryParam(r.URL.Query(), "status", "active")

// Extract Bearer token from Authorization header
token, ok := request.ExtractBearerToken(r)

// Check Content-Type
if request.IsJSON(r) { /* decode JSON body */ }

// Require a header or return error
tenantID, err := request.RequireHeader(r, "X-Tenant-ID")

// Tri-state boolean query param (nil when absent)
active := request.OptionalQueryBool(r.URL.Query(), "active")
```

---

### health — Health Check Handlers

```go
import "github.com/Saver-Street/cat-shared-lib/health"

h := health.Handler("billing-service", "1.2.0",
    health.DBChecker(pool),
    health.NewChecker("redis", func(ctx context.Context) error {
        return redisClient.Ping(ctx).Err()
    }),
    health.NewChecker("stripe", func(ctx context.Context) error {
        return stripeClient.Ping(ctx)
    }),
)
mux.Handle("/health", h)
// GET /health → {"status":"ok","service":"billing-service","checks":{"db":"ok","redis":"ok","stripe":"ok"}}
// Returns 503 if any checker fails.
```

---

### httpclient — Resilient HTTP Client

The `httpclient` package wraps `net/http` with automatic retries, exponential
backoff with jitter, optional circuit breaker integration, and JSON convenience
methods.

```go
import (
    "github.com/Saver-Street/cat-shared-lib/httpclient"
    "github.com/Saver-Street/cat-shared-lib/circuitbreaker"
)

cb := circuitbreaker.New("payments-api",
    circuitbreaker.WithFailureThreshold(5),
    circuitbreaker.WithSuccessThreshold(2),
    circuitbreaker.WithOpenTimeout(30*time.Second),
)

client := httpclient.New(
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithRetries(3),
    httpclient.WithBaseBackoff(200 * time.Millisecond),
    httpclient.WithMaxBackoff(5 * time.Second),
    httpclient.WithCircuitBreaker(cb),
    httpclient.WithHeader("X-Service", "billing"),
    httpclient.WithUserAgent("billing-service/1.0"),
    httpclient.WithRequestHook(func(req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+getToken())
        return nil
    }),
)

// Typed JSON helpers
var result PaymentResponse
err := client.GetJSON(ctx, "https://api.payments.io/v1/payment/123", &result)
err  = client.PostJSON(ctx, "https://api.payments.io/v1/payments", &payload, &result)
err  = client.PutJSON(ctx, "https://api.payments.io/v1/payment/123", &update, &result)
err  = client.DeleteJSON(ctx, "https://api.payments.io/v1/payment/123", nil)

// Raw request
resp, err := client.Do(ctx, http.MethodGet, url, nil)
fmt.Println(resp.StatusCode, string(resp.Body))
```

---

### apperror — Structured Errors

```go
import "github.com/Saver-Street/cat-shared-lib/apperror"

// Construct typed errors
err := apperror.NotFound("user not found")
err  = apperror.BadRequest("invalid email format")
err  = apperror.Unauthorized("token expired")
err  = apperror.Forbidden("not an admin")
err  = apperror.InternalWrap("query failed", dbErr) // wraps original

// In handlers: get HTTP status from any error
if appErr, ok := err.(*apperror.Error); ok {
    http.Error(w, appErr.Message, appErr.Status)
    return
}
// Or use the helper:
status := apperror.HTTPStatus(err) // 404, 400, 401, 403, 500, …
```

---

### Additional Packages

**`circuitbreaker`** — Protects downstream calls from cascading failures.

```go
cb := circuitbreaker.New("user-service",
    circuitbreaker.WithFailureThreshold(5),
    circuitbreaker.WithOpenTimeout(30*time.Second),
)
err := cb.Execute(func() error { return callUserService(ctx) })
if errors.Is(err, circuitbreaker.ErrCircuitOpen) { /* fast-fail */ }
```

**`ratelimit`** — Per-key sliding-window + token-bucket limiter.

```go
rl := ratelimit.New(ratelimit.Config{RequestsPerSecond: 10, Burst: 20})
if !rl.Allow(r.RemoteAddr) {
    response.TooManyRequests(w)
    return
}
```

**`cors`** — CORS middleware.

```go
mux.Handle("/", cors.Middleware(cors.Config{
    AllowedOrigins: []string{"https://app.example.com"},
    AllowedMethods: []string{"GET","POST","PUT","DELETE"},
})(handler))
```

**`featureflags`** — Toggle features with environment variables.

```go
if featureflags.Enabled("NEW_BILLING_FLOW") {
    return newBillingHandler(w, r)
}
```

**`testkit`** — Lightweight test helpers and assertion library.

```go
// Assert* helpers use t.Errorf (non-fatal):
testkit.AssertEqual(t, got, want)
testkit.AssertNoError(t, err)
testkit.AssertError(t, err)
testkit.AssertErrorContains(t, err, "not found")
testkit.AssertContains(t, body, "success")
testkit.AssertStatus(t, rr, http.StatusOK)
testkit.AssertGreater(t, elapsed, 0)
testkit.AssertLess(t, latency, maxAllowed)
testkit.AssertHasPrefix(t, path, "/api/")
testkit.AssertHasSuffix(t, file, ".json")
testkit.AssertMapHasKey(t, headers, "Content-Type")
testkit.AssertWithin(t, elapsed, 100*time.Millisecond)

// Require* helpers use t.Fatalf (fatal — stops test on failure):
testkit.RequireNoError(t, err)   // guard: stop test if err is non-nil
testkit.RequireEqual(t, got, want)
testkit.RequireNotNil(t, obj)    // guard: stop test if obj is nil
testkit.RequireLen(t, items, 3)  // guard: stop test if wrong length

// Mock server for HTTP client tests:
ms := testkit.NewMockServer(t)
ms.Handle(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]any{"ok": true})
})
client.GetJSON(ctx, ms.URL+"/path", &result)
testkit.AssertEqual(t, ms.RequestCount(), 1)
```

**`security`** — Input validation, PII redaction, and security helpers.

```go
security.ContainsSuspiciousInput(input)    // SQL/XSS detection
security.RedactPII(data)                   // mask PII fields in maps
security.SanitizeHeader(s)                 // strip CRLF from headers
security.IsRelativeURL("/dashboard")       // safe redirect check
security.MaskEmail("alice@example.com")    // → "a****@example.com"
security.RedactURL("https://u:p@host/db") // → "https://REDACTED@host/db"
security.SanitizeFilename("../../etc/passwd") // → "passwd"
security.CSPHeader(map[string]string{          // build CSP header
    "default-src": "'self'",
    "script-src":  "'self' 'unsafe-inline'",
})
```

**`sanitize`** — Input sanitization and string processing.

```go
sanitize.StripHTML("<b>Hello</b>")         // → "Hello"
sanitize.EscapeHTML("<script>")            // → "&lt;script&gt;"
sanitize.Mask("sk-abc123def456", 6)        // → "*********def456"
sanitize.Slugify("My Blog Post!")          // → "my-blog-post"
sanitize.Truncate("long string", 8)       // → "long st…"
sanitize.TrimStrings([]string{" a ", ""})  // → ["a"]
sanitize.NormalizeWhitespace("a  b\tc")   // → "a b c"
sanitize.RemoveNonPrintable("a\x00b")     // → "ab"
sanitize.CamelToSnake("HTTPClient")       // → "http_client"
sanitize.SnakeToCamel("http_client")      // → "httpClient"
sanitize.Deref(ptr, "default")            // generic pointer dereference
```

---

### contracts — Service Interface Contracts

The `contracts` package defines the Go interfaces that every microservice in the
platform must implement, providing compile-time enforcement of service contracts
and a canonical JSON error body type.

```go
import "github.com/Saver-Street/cat-shared-lib/contracts"

// Implement the full Service contract in your service type:
//
//   type BillingService struct { ... }
//   func (s *BillingService) Name() string        { return "billing-service" }
//   func (s *BillingService) Version() string     { return "1.0.0" }
//   func (s *BillingService) Environment() string { return os.Getenv("ENV") }
//   func (s *BillingService) RegisterRoutes(mux *http.ServeMux) { ... }
//   func (s *BillingService) HealthCheck(ctx context.Context) (contracts.HealthStatus, error) { ... }
//
// var _ contracts.Service = (*BillingService)(nil) // compile-time check

// Canonical JSON error body for all non-2xx responses:
errBody := contracts.NewStandardError("NOT_FOUND", "user not found")
errBody = contracts.NewStandardErrorWithDetails("VALIDATION_ERROR", "invalid input",
    map[string]any{"field": "email", "reason": "invalid format"},
)
json.NewEncoder(w).Encode(errBody)

// Health status:
status := contracts.HealthStatus{
    State:   contracts.HealthStateOK,
    Service: "billing-service",
    Version: "1.0.0",
    Checks:  map[string]string{"db": "ok", "cache": "ok"},
}
```

---

### servicetest — Integration Test Helpers

The `servicetest` package provides HTTP test server routing, request recording,
a mock row/DB helper for database-free unit tests, and an in-memory fixture
registry.

```go
import "github.com/Saver-Street/cat-shared-lib/servicetest"

// --- HTTP test server with per-route stubs ---
srv := servicetest.NewHTTPTestServer(t) // closes automatically via t.Cleanup

srv.Handle(http.MethodGet, "/users", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode([]map[string]any{{"id": "1", "email": "a@example.com"}})
})

// Stub a JSON response directly:
srv.HandleJSON(http.MethodPost, "/users", http.StatusCreated, map[string]any{"id": "2"})

// Inspect recorded requests:
reqs := srv.Requests()
last := srv.LastRequest() // *RecordedRequest{Method, Path, Headers, Body, Query}

srv.Reset() // clear recorded requests between sub-tests

// --- Mock row / DB helper (no real DB required) ---
row := &servicetest.MockRow{}
row.Set([]any{"uuid-1", "alice@example.com"}) // values scanned in order

db := &servicetest.DBTestHelper{}
db.QueueRow(row)
// Pass db wherever a database.Querier is expected

// Inspect executed queries:
fmt.Println(db.QueryCount())
for _, q := range db.Queries() {
    fmt.Println(q.SQL, q.Args)
}

// --- In-memory fixture registry ---
fixtures := servicetest.NewFixtures()
fixtures.RegisterJSON("user", map[string]any{"id": "1", "email": "a@example.com"})

raw := fixtures.MustLoad("user")
var u User
fixtures.LoadInto("user", &u)
```

---

## Querier Interface

DB-querying functions accept a `Querier` interface rather than a concrete pool,
making them work identically with `*pgxpool.Pool`, `*pgx.Conn`, and `pgx.Tx`:

```go
type Querier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
```

---

## Design Notes

- **DirectDB only** in Phase C — no HTTP service-to-service calls via service packages
- **Boolean flags are plain-text** (`"true"` / `"false"`) — no encryption
- **Rate limiter** is per-IP sliding window + token bucket, safe for concurrent use
- **Zero-config defaults** — all packages work out of the box with sensible defaults
- **Context-aware** — every long-running operation accepts `context.Context`

---

## Testing

Every package includes multiple layers of testing:

| Layer | Files | Description |
|-------|-------|-------------|
| Unit tests | `*_test.go` | ≥95% statement coverage enforced by CI |
| Example tests | 34 `example_test.go` | Runnable godoc examples for every package |
| Fuzz tests | 21 `fuzz_test.go` | Robustness against arbitrary input |
| Benchmarks | 17 `bench_test.go` | Performance regression detection |

```sh
make test            # Run all unit tests
make test-race       # Run tests with Go race detector
make bench           # Run all benchmarks
make fuzz            # Smoke-run every fuzz target (5s each)
make cover           # Generate coverage report
make check-coverage  # Verify all packages meet 95% threshold
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for testing guidelines and examples.

