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
make test       # Run unit tests
make test-v     # Run tests (verbose)
make test-race  # Run tests with race detector
make lint       # Run linters (go vet + staticcheck)
make cover      # Generate coverage report
```

### Package Structure
```
entitlements/   # Subscription tier limits + DirectDB queries
flags/          # Feature flag DirectDB reads
identity/       # Candidate resolution + context getters
middleware/     # JWT context, rate limiting, brute-force protection
request/        # HTTP request parsing, URL param extraction
response/       # JSON response helpers
sanitize/       # Filename sanitization, NilIfEmpty, IsDuplicateKey
scan/           # Generic database row scanning
security/       # Input validation, PII redaction
types/          # Shared domain types (User, CandidateProfile, Pagination)
```

### Key Design Principles
- **Querier interface**: DB functions accept `Querier` (pool, conn, or tx)
- **100% test coverage**: All packages maintain 100% coverage
- **No HTTP calls**: DirectDB only in Phase C — no service-to-service HTTP
- **Boolean flags**: Stored as plain-text (`"true"` / `"false"`)

### Adding a Package
1. Create directory with descriptive name
2. Add package with exported types/functions
3. Write comprehensive tests (target 100% coverage)
4. Update README.md package table
5. Tag new version after merge

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
