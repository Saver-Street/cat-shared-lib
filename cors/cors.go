// Package cors provides configurable Cross-Origin Resource Sharing (CORS)
// middleware for HTTP servers.
package cors

import (
	"net/http"
	"strconv"
	"strings"
)

// Config holds CORS policy settings.
type Config struct {
	// AllowedOrigins is the list of origins permitted to make requests.
	// Use ["*"] to allow any origin. Default: ["*"].
	AllowedOrigins []string
	// AllowedMethods is the list of HTTP methods permitted.
	// Default: [GET, POST, PUT, PATCH, DELETE, OPTIONS].
	AllowedMethods []string
	// AllowedHeaders is the list of headers clients may send.
	// Default: [Content-Type, Authorization, X-Request-ID].
	AllowedHeaders []string
	// ExposedHeaders is the list of headers the browser may access.
	ExposedHeaders []string
	// AllowCredentials indicates whether cookies/auth headers are allowed.
	AllowCredentials bool
	// MaxAge is how long (in seconds) the preflight result can be cached.
	// Default: 86400 (24 hours).
	MaxAge int
}

func (c *Config) defaults() {
	if len(c.AllowedOrigins) == 0 {
		c.AllowedOrigins = []string{"*"}
	}
	if len(c.AllowedMethods) == 0 {
		c.AllowedMethods = []string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, http.MethodOptions,
		}
	}
	if len(c.AllowedHeaders) == 0 {
		c.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	if c.MaxAge <= 0 {
		c.MaxAge = 86400
	}
}

// Middleware returns an HTTP middleware that applies the CORS policy.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	cfg.defaults()

	allowAllOrigins := len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*"
	originSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		originSet[strings.ToLower(o)] = struct{}{}
	}

	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")
	exposed := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowed := allowAllOrigins
			if !allowed {
				_, allowed = originSet[strings.ToLower(origin)]
			}
			if !allowed {
				next.ServeHTTP(w, r)
				return
			}

			// Set the allowed origin.
			if allowAllOrigins && !cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if exposed != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposed)
			}

			// Handle preflight.
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				w.Header().Set("Access-Control-Max-Age", maxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
