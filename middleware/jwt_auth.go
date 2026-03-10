// Package middleware provides HTTP middleware for authentication, authorization,
// rate limiting, and brute-force protection used across microservices.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// JWTClaims holds the standard and custom claims extracted from a JWT token.
type JWTClaims struct {
	// Subject is the user identifier (standard "sub" claim).
	Subject string `json:"sub"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Role is the user's role (e.g. "admin", "user").
	Role string `json:"role"`
	// IssuedAt is the token creation time (standard "iat" claim).
	IssuedAt int64 `json:"iat"`
	// ExpiresAt is the token expiration time (standard "exp" claim).
	ExpiresAt int64 `json:"exp"`
	// Issuer is the entity that issued the token (standard "iss" claim).
	Issuer string `json:"iss"`
}

// Sentinel errors returned by JWT validation.
var (
	ErrMissingToken   = errors.New("middleware: missing authorization token")
	ErrInvalidToken   = errors.New("middleware: invalid token format")
	ErrInvalidAlg     = errors.New("middleware: unsupported signing algorithm")
	ErrSignatureFail  = errors.New("middleware: signature verification failed")
	ErrTokenExpired   = errors.New("middleware: token has expired")
	ErrMissingSubject = errors.New("middleware: token missing subject claim")
)

// JWTAuthConfig configures the JWT authentication middleware.
type JWTAuthConfig struct {
	// Secret is the HMAC-SHA256 signing key. Must not be empty.
	Secret []byte
	// Issuer, if non-empty, is validated against the token's "iss" claim.
	Issuer string
	// SkipPaths are URL paths that bypass authentication (e.g. "/health").
	SkipPaths []string
	// NowFunc returns the current time; defaults to time.Now.
	// Useful for testing.
	NowFunc func() time.Time
}

// now returns the current time, using NowFunc if set.
func (c *JWTAuthConfig) now() time.Time {
	if c.NowFunc != nil {
		return c.NowFunc()
	}
	return time.Now()
}

// JWTAuth returns HTTP middleware that validates HS256-signed JWT tokens
// from the Authorization header (Bearer scheme). On success it populates
// the request context with the user ID, email, and role via SetUserID,
// SetUserEmail, and SetUserRole.
func JWTAuth(cfg JWTAuthConfig) func(http.Handler) http.Handler {
	skipSet := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipSet[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := validateJWT(r, cfg.Secret, cfg.Issuer, cfg.now())
			if err != nil {
				status := http.StatusUnauthorized
				response.Error(w, status, err.Error())
				return
			}

			ctx := r.Context()
			ctx = SetUserID(ctx, claims.Subject)
			if claims.Email != "" {
				ctx = SetUserEmail(ctx, claims.Email)
			}
			if claims.Role != "" {
				ctx = SetUserRole(ctx, claims.Role)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateJWT extracts and validates a JWT token from the request's
// Authorization header. It supports only HS256 (HMAC-SHA256) tokens.
func validateJWT(r *http.Request, secret []byte, issuer string, now time.Time) (*JWTClaims, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, ErrMissingToken
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return nil, ErrInvalidToken
	}
	tokenStr := auth[len(prefix):]

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Decode and validate header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}
	if header.Alg != "HS256" {
		return nil, ErrInvalidAlg
	}

	// Verify signature
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrSignatureFail
	}
	if !hmac.Equal(expectedSig, actualSig) {
		return nil, ErrSignatureFail
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Validate expiration
	if claims.ExpiresAt > 0 && now.Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	// Validate issuer
	if issuer != "" && claims.Issuer != issuer {
		return nil, ErrInvalidToken
	}

	// Validate subject
	if claims.Subject == "" {
		return nil, ErrMissingSubject
	}

	return &claims, nil
}

// SignHS256 creates an HS256-signed JWT token from the given claims.
// This is a convenience for services that issue their own tokens.
func SignHS256(claims JWTClaims, secret []byte) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	sigInput := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return sigInput + "." + sig, nil
}
