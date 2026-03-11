package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSConfig configures Cross-Origin Resource Sharing (CORS) headers.
type CORSConfig struct {
	// AllowedOrigins is the list of origins that are permitted. Use "*" to
	// allow all origins (not recommended for credentialed requests).
	AllowedOrigins []string
	// AllowedMethods lists the HTTP methods allowed for cross-origin requests.
	// Default: GET, POST, PUT, PATCH, DELETE, OPTIONS.
	AllowedMethods []string
	// AllowedHeaders lists the headers the client may send in requests.
	// Default: Content-Type, Authorization, X-Request-ID.
	AllowedHeaders []string
	// ExposedHeaders lists the headers the browser may expose to JavaScript.
	ExposedHeaders []string
	// AllowCredentials indicates whether cookies and credentials are allowed.
	AllowCredentials bool
	// MaxAge is how long the preflight response can be cached.
	// Default: 5 minutes.
	MaxAge time.Duration
}

func (c *CORSConfig) defaults() {
	if len(c.AllowedMethods) == 0 {
		c.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(c.AllowedHeaders) == 0 {
		c.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	if c.MaxAge <= 0 {
		c.MaxAge = 5 * time.Minute
	}
}

// CORS returns middleware that handles CORS preflight and standard requests.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	cfg.defaults()

	allowedOriginSet := make(map[string]bool, len(cfg.AllowedOrigins))
	allowAll := false
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			allowAll = true
		}
		allowedOriginSet[o] = true
	}

	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")
	exposed := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !allowAll && !allowedOriginSet[origin] {
				next.ServeHTTP(w, r)
				return
			}

			h := w.Header()
			if allowAll && !cfg.AllowCredentials {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Vary", "Origin")
			}

			if cfg.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if exposed != "" {
				h.Set("Access-Control-Expose-Headers", exposed)
			}

			if r.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", methods)
				h.Set("Access-Control-Allow-Headers", headers)
				h.Set("Access-Control-Max-Age", maxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
