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
| `middleware` | JWT auth, request ID, logging, recovery, rate limiting, ETag, caching, basic auth, compression | 99.7% |
| `config` | Env var parsing with defaults, validation, byte sizes, URLs, ports, enums | 100% |
| `database` | PostgreSQL connection pool, transaction helpers | 96.6% |
| `validation` | Email, UUID, phone, URL, slug, alphanumeric, hex, numeric, range, IP, JSON, base64 | 100% |
| `cache` | Generic in-memory LRU cache with per-entry TTL | 98% |
| `retry` | Exponential backoff with jitter and context cancellation | 100% |
| `crypto` | bcrypt password hashing, secure tokens, HMAC-SHA256, SHA-256, random strings | 100% |
| `email` | SMTP mailer with HTML/text template support | 92.6% |
| `tracing` | OpenTelemetry distributed tracing setup and helpers | 98.5% |
| `migration` | Database migration runner with rollback support | 100% |
| `response` | JSON/XML/HTML response helpers, pagination, SSE, streaming, redirects | 100% |
| `request` | HTTP request parsing, URL params, JSON body decoding, ID lists | 100% |
| `health` | Standardized health check handlers with concurrent checkers | 99.2% |
| `httpclient` | Resilient HTTP client with retries, circuit breaker, HEAD support | 100% |
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
| `sanitize` | Filename sanitization, HTML escaping, generics (Unique, Filter, Map, Compact, Contains) | 100% |
| `scan` | Generic database row scanning (Rows, Row, First) | 100% |
| `security` | Input detection, PII redaction, password strength, CSP, URL scrubbing | 100% |
| `server` | HTTP server with graceful shutdown (SIGINT/SIGTERM) | 100% |
| `shutdown` | OS signal-based graceful shutdown coordinator | 100% |
| `testkit` | 40+ assertion helpers, mock server, call recorder, RequireTrue/RequireFalse | 100% |
| `types` | Shared domain types (User, CandidateProfile, Pagination) | 100% |
| `contracts` | Shared service interfaces (Service, Handler, HealthCheck, StandardError) | 100% |
| `servicetest` | Integration test helpers: HTTP test server, mock Querier, fixture loader | 100% |
| `randutil` | Random utilities: Pick, Shuffle, Sample, WeightedPick, string generators | 100% |
| `stringutil` | String utilities: case conversion, padding, word wrap, blank check | 100% |
| `timeout` | Timeout utilities: Do, DoSimple, After, Race with deadlines | 100% |
| `pubsub` | Typed, in-process publish/subscribe event bus | 100% |
| `worker` | Bounded, context-aware worker pool for concurrent job processing | 100% |
| `schedule` | Lightweight in-process periodic task scheduler with named tasks | 100% |

---


## Usage & Reference

Detailed usage examples are split into focused guides:

- **[Core Packages Guide](./PACKAGES-GUIDE.md)** — middleware, config, database, validation, cache, retry, crypto, email, tracing, migration
- **[Advanced Packages Guide](./PACKAGES-ADVANCED.md)** — response, request, health, httpclient, apperror, contracts, servicetest
- **[Usage Quick Start](./USAGE.md)** — Getting started, common patterns
- **[Migration Guide](./MIGRATION.md)** — Adopting the shared library

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