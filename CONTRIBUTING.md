# Contributing to cat-shared-lib

## Quick Start

```bash
git clone https://github.com/Saver-Street/cat-shared-lib.git
cd cat-shared-lib
go mod download
go test ./...
```

## Development

### Prerequisites
- Go 1.25+

### Build & Test
```bash
make test            # Run unit tests
make test-v          # Run tests (verbose)
make test-race       # Run tests with race detector
make lint            # Run linters (go vet + staticcheck)
make cover           # Generate coverage report
make cover-html      # Generate HTML coverage report
make check-coverage  # Verify all packages meet 95% threshold
make bench           # Run all benchmarks
make fuzz            # Run all fuzz tests (5s smoke run per target)
```

### Package Structure
```
apperror/        # Structured application errors with HTTP status codes
cache/           # Generic in-memory cache with TTL and LRU eviction
circuitbreaker/  # Circuit breaker for external service calls
config/          # Configuration loading from environment variables
contracts/       # Shared interfaces and types for service contracts
cors/            # CORS middleware for HTTP servers
crypto/          # Password hashing, HMAC, secure token generation
database/        # PostgreSQL pool setup and transaction helpers
discovery/       # Service discovery and health-check registration
email/           # SMTP mailer with HTML/text template support
entitlements/    # Subscription tier limits + DirectDB queries
featureflags/    # Feature flag evaluation
flags/           # Feature flag DirectDB reads
health/          # Health-check HTTP handler and dependency checks
httpclient/      # HTTP client with retries, backoff, circuit breaker
identity/        # Candidate resolution + context getters
metrics/         # Prometheus-style metrics (counters, histograms)
middleware/      # JWT auth, rate limiting, logging, recovery
migration/       # Lightweight database migration runner
openapi/         # OpenAPI spec builder
ratelimit/       # Sliding-window rate limiter
request/         # HTTP request parsing, URL param extraction
response/        # JSON response helpers
retry/           # Retry with exponential backoff and jitter
sanitize/        # Filename sanitization, NilIfEmpty, IsDuplicateKey
scan/            # Generic database row scanning
security/        # Input validation, PII redaction
server/          # HTTP server with graceful shutdown defaults
servicetest/     # Integration test helpers (HTTP, DB mocks, fixtures)
shutdown/        # Graceful shutdown with signal handling and draining
testkit/         # Assertion helpers and mock utilities for tests
tracing/         # OpenTelemetry distributed tracing setup
types/           # Shared domain types (User, CandidateProfile, Pagination)
validation/      # Field validation (email, UUID, phone, URL)
```

### Key Design Principles
- **Querier interface**: DB functions accept `Querier` (pool, conn, or tx)
- **95% test coverage**: Enforced by `make check-coverage`; target 100% where feasible
- **No HTTP calls**: DirectDB only in Phase C — no service-to-service HTTP
- **Boolean flags**: Stored as plain-text (`"true"` / `"false"`)

### Adding a Package
1. Create directory with descriptive name
2. Add package with exported types/functions
3. Add a `doc.go` with a package-level godoc comment
4. Write comprehensive tests (target 100% coverage)
5. Add `example_test.go` with runnable godoc examples
6. Add `fuzz_test.go` for packages that parse or validate input
7. Add `bench_test.go` for performance-critical functions
8. Update README.md package table
9. Tag new version after merge

### Testing Strategy

This library uses a multi-tier testing approach:

| Test Type | File | Purpose | When to Add |
|-----------|------|---------|-------------|
| Unit tests | `*_test.go` | Correctness, edge cases, error paths | Every package (≥95% coverage) |
| Example tests | `example_test.go` | Living documentation via `godoc` | Every package |
| Fuzz tests | `fuzz_test.go` | Robustness against arbitrary input | Parsers, validators, crypto, serializers |
| Benchmarks | `bench_test.go` | Performance regression detection | Hot-path functions (middleware, parsing, encoding) |

**Fuzz tests** use Go's built-in fuzzing (`go test -fuzz`). Seed with representative
inputs and verify the function never panics:

```go
func FuzzMyParser(f *testing.F) {
    f.Add("valid input")
    f.Add("")
    f.Fuzz(func(t *testing.T, input string) {
        _ = MyParser(input) // must not panic
    })
}
```

**Benchmarks** use `b.Loop()` (Go 1.25+) for iteration:

```go
func BenchmarkMyFunc(b *testing.B) {
    for b.Loop() {
        MyFunc("input")
    }
}
```

### Code Style
- Follow standard Go conventions (`gofmt`, `go vet`)
- CI enforces `gofmt` — run `gofmt -w .` before committing
- Use `context.Context` as first parameter for DB functions
- Export types that services need, keep internals unexported

## Versioning

This library uses semantic versioning. After merging changes:

```bash
git tag v1.X.0
git push origin v1.X.0
```

Consumers update with: `go get github.com/Saver-Street/cat-shared-lib@v1.X.0`
