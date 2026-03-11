// Package crypto provides cryptographic utilities for password hashing, secure
// token generation, HMAC signing, AES-GCM encryption, and constant-time
// comparison.
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
// and [EqualBytes] provide constant-time comparison to prevent timing
// side-channel attacks.
//
// [Encrypt] and [Decrypt] provide AES-GCM authenticated encryption for
// protecting data at rest.  [EncryptString] and [DecryptString] are
// convenience wrappers that produce URL-safe base64 output.
package crypto
