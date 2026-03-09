# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. Email security concerns to the repository maintainers
3. Include steps to reproduce and potential impact

We aim to acknowledge reports within 48 hours and provide a fix timeline within 7 days.

## Security Features

### Input Validation
- PII redaction (`security.RedactPII`) for safe logging
- Filename sanitization (path traversal prevention)
- Input length validation helpers

### Rate Limiting
- Per-IP sliding window rate limiter
- Memory-bounded (configurable max entries with periodic sweep)
- Brute-force protection middleware

### Database
- Querier interface — prevents accidental raw query construction
- Parameterized queries only (no string concatenation)

### Middleware
- JWT context extraction and validation
- Role-based access control helpers

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 1.4.x   | ✅        |
| < 1.4   | ❌        |
