# Advanced Packages Guide

> Extracted from the main README. See also: [README](./README.md) | [Core Packages](./PACKAGES-GUIDE.md) | [Usage Guide](./USAGE.md)


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


---

## Related Documentation

- [README](./README.md) — Overview & installation
- [Core Packages](./PACKAGES-GUIDE.md) — Middleware, config, database, validation, cache, retry
- [Usage Guide](./USAGE.md) — Quick start & common patterns
- [Migration Guide](./MIGRATION.md) — Adoption steps
