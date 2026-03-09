# cat-shared-lib

Shared Go library for Catherine (Auto-Apps) microservices.

**Wave 0 — required by all service extractions.**

## Packages

| Package | Description |
|---|---|
| `entitlements` | Subscription tier limits + DirectDB queries |
| `flags` | Feature flag DirectDB reads (boolean flags, plain-text) |
| `identity` | Candidate resolution + context getters |
| `types` | Shared domain types (User, CandidateProfile, Pagination) |
| `response` | JSON response helpers |
| `middleware` | JWT context, rate limiting, brute-force protection |

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

## Design Notes

- **DirectDB only** in Phase C — no HTTP service-to-service calls
- **Boolean flags are plain-text** (`"true"` / `"false"`) — no encryption
- **Rate limiter** is per-IP sliding window, safe for concurrent use
- All packages have 100% test coverage on pure functions

## Import in services

```go
// go.mod
require github.com/Saver-Street/cat-shared-lib v1.0.0
```
