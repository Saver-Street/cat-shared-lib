package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

// CSRFConfig controls the behavior of the [CSRF] middleware.
type CSRFConfig struct {
	// TokenLength is the byte length of generated tokens (hex-encoded to
	// double this length in headers/forms). Default is 32.
	TokenLength int

	// TokenHeader is the HTTP header the client sends the token in.
	// Default is "X-CSRF-Token".
	TokenHeader string

	// FormField is the form field name to check when the header is absent.
	// Default is "csrf_token".
	FormField string

	// CookieName is the name of the cookie that stores the token.
	// Default is "_csrf".
	CookieName string

	// CookiePath sets the Path attribute of the CSRF cookie.
	// Default is "/".
	CookiePath string

	// Secure sets the Secure flag on the cookie. Default is true.
	Secure *bool

	// SameSite sets the SameSite attribute. Default is [http.SameSiteStrictMode].
	SameSite http.SameSite

	// ErrorHandler is called when a CSRF check fails. If nil the middleware
	// responds with 403 Forbidden.
	ErrorHandler http.Handler

	// SkipCheck returns true for requests that should bypass CSRF validation
	// (e.g., public endpoints). Safe methods (GET, HEAD, OPTIONS, TRACE) are
	// always skipped regardless of this function.
	SkipCheck func(r *http.Request) bool
}

func (c *CSRFConfig) defaults() {
	if c.TokenLength <= 0 {
		c.TokenLength = 32
	}
	if c.TokenHeader == "" {
		c.TokenHeader = "X-CSRF-Token"
	}
	if c.FormField == "" {
		c.FormField = "csrf_token"
	}
	if c.CookieName == "" {
		c.CookieName = "_csrf"
	}
	if c.CookiePath == "" {
		c.CookiePath = "/"
	}
	if c.Secure == nil {
		t := true
		c.Secure = &t
	}
	if c.SameSite == 0 {
		c.SameSite = http.SameSiteStrictMode
	}
}

// CSRFTokenKey is the context key used to store the CSRF token.
const CSRFTokenKey contextKey = "csrfToken"

// GetCSRFToken extracts the CSRF token from the request context.
func GetCSRFToken(r *http.Request) string {
	v, _ := r.Context().Value(CSRFTokenKey).(string)
	return v
}

// csrfRandRead is the random byte source, overridable in tests.
var csrfRandRead = rand.Read

// CSRF returns middleware that protects against cross-site request forgery.
//
// On every request the middleware ensures a CSRF cookie exists (creating one if
// needed) and stores the token in the request context (retrieve it with
// [GetCSRFToken]). For unsafe HTTP methods (POST, PUT, PATCH, DELETE) the
// middleware validates that the token submitted via header or form field matches
// the cookie value using constant-time comparison.
func CSRF(cfg CSRFConfig) func(http.Handler) http.Handler {
	cfg.defaults()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read or generate token.
			token := csrfTokenFromCookie(r, cfg.CookieName)
			if token == "" {
				token = csrfGenerateToken(cfg.TokenLength)
			}

			// Always set the cookie so the expiry is refreshed.
			http.SetCookie(w, &http.Cookie{
				Name:     cfg.CookieName,
				Value:    token,
				Path:     cfg.CookiePath,
				HttpOnly: true,
				Secure:   *cfg.Secure,
				SameSite: cfg.SameSite,
			})

			// Store in context for handlers/templates.
			r = r.WithContext(context.WithValue(r.Context(), CSRFTokenKey, token))

			// Safe methods and skipped paths bypass validation.
			if csrfIsSafeMethod(r.Method) || (cfg.SkipCheck != nil && cfg.SkipCheck(r)) {
				next.ServeHTTP(w, r)
				return
			}

			// Validate submitted token.
			submitted := r.Header.Get(cfg.TokenHeader)
			if submitted == "" {
				submitted = r.FormValue(cfg.FormField)
			}

			if !csrfTokensMatch(token, submitted) {
				if cfg.ErrorHandler != nil {
					cfg.ErrorHandler.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Forbidden - CSRF token invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func csrfTokenFromCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func csrfGenerateToken(length int) string {
	b := make([]byte, length)
	_, _ = csrfRandRead(b)
	return hex.EncodeToString(b)
}

func csrfTokensMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func csrfIsSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}
