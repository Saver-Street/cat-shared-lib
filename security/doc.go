// Package security provides input validation and PII redaction utilities for
// protecting Catherine microservices against injection attacks and data leaks.
//
// [ContainsSuspiciousInput] checks a string for common SQL injection and XSS
// patterns (e.g. UNION SELECT, <script>, javascript: URIs) and returns true if
// any are detected.
//
// [RedactPII] recursively walks a map and replaces values whose keys match
// sensitive field names (password, token, ssn, etc.) with a "[REDACTED]"
// placeholder, making the result safe for audit logging.
//
// [TruncateForLog] shortens a string to a maximum length and strips control
// characters, preventing log injection and unbounded log entries.
package security
