// Package auth provides JWT validation utilities for Cat microservices.
//
// This package allows services to validate auth-service JWTs locally using
// a shared HS256 secret, without making HTTP calls to auth-service. This is
// faster (~5μs vs ~5ms) and doesn't require auth-service to be running.
//
// Usage:
//
//	validator, err := auth.NewLocalValidator(os.Getenv("SERVICE_JWT_SECRET"))
//	if err != nil { log.Fatal(err) }
//	claims, err := validator.Validate(tokenString)
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the decoded JWT claims from auth-service tokens.
type Claims struct {
	UserID        string `json:"uid"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	IsTester      bool   `json:"is_tester"`
	EmailVerified bool   `json:"email_verified"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	ExpiresIn     int    `json:"expires_in,omitempty"`
}

var (
	ErrUnauthorized = errors.New("auth: token is invalid or expired")
	ErrSecretTooShort = errors.New("auth: JWT secret must be at least 32 bytes")
)

// LocalValidator validates JWTs locally using the shared HS256 secret.
type LocalValidator struct {
	secret []byte
}

// NewLocalValidator creates a validator using the shared JWT secret.
// The secret must match AUTH_JWT_SECRET on the auth-service.
func NewLocalValidator(secret string) (*LocalValidator, error) {
	if len(secret) < 32 {
		return nil, ErrSecretTooShort
	}
	return &LocalValidator{secret: []byte(secret)}, nil
}

// Validate parses and validates a JWT token locally.
// Returns claims if the token is valid, or ErrUnauthorized if not.
func (v *LocalValidator) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if !token.Valid {
		return nil, ErrUnauthorized
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	claims := &Claims{
		UserID:        claimStr(mapClaims, "uid"),
		Email:         claimStr(mapClaims, "email"),
		Name:          claimStr(mapClaims, "name"),
		Role:          claimStr(mapClaims, "role"),
		Status:        claimStr(mapClaims, "status"),
		IsTester:      claimBool(mapClaims, "is_tester"),
		EmailVerified: claimBool(mapClaims, "email_verified"),
	}

	if exp, err := mapClaims.GetExpirationTime(); err == nil && exp != nil {
		claims.ExpiresAt = exp.Format(time.RFC3339)
		claims.ExpiresIn = int(time.Until(exp.Time).Seconds())
		if claims.ExpiresIn < 0 {
			claims.ExpiresIn = 0
		}
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf("%w: missing uid claim", ErrUnauthorized)
	}

	return claims, nil
}

func claimStr(m jwt.MapClaims, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func claimBool(m jwt.MapClaims, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
