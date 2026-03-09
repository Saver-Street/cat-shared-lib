# cat-shared-lib

Shared Go library for Catherine (Auto-Apps) microservices.

**Wave 0 — required by all service extractions.**

## Packages

| Package | Description | Coverage |
|---|---|---|
| `entitlements` | Subscription tier limits + DirectDB queries | 100% |
| `flags` | Feature flag DirectDB reads (boolean flags, plain-text) | 100% |
| `identity` | Candidate resolution + context getters | 100% |
| `types` | Shared domain types (User, CandidateProfile, Pagination) | 100% |
| `response` | JSON response helpers | 100% |
| `middleware` | JWT context, rate limiting, brute-force protection | 100% |

## Usage

```go
import (
    "github.com/Saver-Street/cat-shared-lib/entitlements"
    "github.com/Saver-Street/cat-shared-lib/flags"
    "github.com/Saver-Street/cat-shared-lib/identity"
    "github.com/Saver-Street/cat-shared-lib/response"
    "github.com/Saver-Street/cat-shared-lib/middleware"
)
```

### Querier Interface

DB-querying functions accept a `Querier` interface instead of concrete `*pgxpool.Pool`:

```go
type Querier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Both `*pgxpool.Pool`, `*pgx.Conn`, and `pgx.Tx` satisfy this interface, making the
library flexible and fully testable without a real database.

```go
// Works with pool
tier, count, err := entitlements.GetUserTierAndUsage(ctx, pool, userID)

// Works with transaction
enabled := flags.IsFeatureEnabled(ctx, tx, "aiScoring")

// Works with single connection
candidateID, err := identity.LookupCandidateID(ctx, conn, userID)
```

## Design Notes

- **DirectDB only** in Phase C — no HTTP service-to-service calls
- **Boolean flags are plain-text** (`"true"` / `"false"`) — no encryption
- **Rate limiter** is per-IP sliding window, safe for concurrent use
- **100% test coverage** across all 6 packages (verified v1.2.0)
- **Querier interface** for DB functions — accepts pool, conn, or tx

## Import in services

```go
// go.mod
require github.com/Saver-Street/cat-shared-lib v1.2.0
```
