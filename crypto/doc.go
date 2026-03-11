// Package crypto provides cryptographic utilities for password hashing, secure
// token generation, HMAC signing, and constant-time comparison.
//
// Passwords are hashed with bcrypt via [HashPassword] (using [DefaultCost]) or
// [HashPasswordWithCost], and verified with [CheckPassword].  [NeedsRehash]
// indicates whether a stored hash should be upgraded to a higher cost factor.
//
// [GenerateToken] and [GenerateHexToken] produce cryptographically random
// strings suitable for session tokens, API keys, and similar secrets.
//
// [HMACSHA256] computes a hex-encoded HMAC-SHA256 signature, and
// [VerifyHMACSHA256] validates one using constant-time comparison.  [Equal]
// provides a general constant-time string comparison to prevent timing
// side-channel attacks.
package crypto
