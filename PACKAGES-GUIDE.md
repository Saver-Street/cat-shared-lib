# Core Packages Guide

> Extracted from the main README. See also: [README](./README.md) | [Advanced Packages](./PACKAGES-ADVANCED.md) | [Usage Guide](./USAGE.md)

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
    Text:    "Hello!\n| \x60maputil\x60 | Generic map utilities: Keys, Values, Merge, Pick, Omit, Filter, Invert | 100% |\n\nYour account is ready.",
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

---

## Related Documentation

- [README](./README.md) — Overview & installation
- [Advanced Packages](./PACKAGES-ADVANCED.md) — Response, health, httpclient, apperror, contracts
- [Usage Guide](./USAGE.md) — Quick start & common patterns
- [Migration Guide](./MIGRATION.md) — Adoption steps
