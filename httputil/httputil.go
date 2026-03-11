package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// IsJSON returns true if the request has a JSON Content-Type.
func IsJSON(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/json")
}

// IsForm returns true if the request has a form-encoded Content-Type.
func IsForm(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/x-www-form-urlencoded")
}

// IsMultipart returns true if the request has a multipart Content-Type.
func IsMultipart(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "multipart/")
}

// IsAJAX returns true if the request has the X-Requested-With: XMLHttpRequest header.
func IsAJAX(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsTLS returns true if the request was made over HTTPS.
func IsTLS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

// Accepts checks if the request accepts the given content type.
func Accepts(r *http.Request, contentType string) bool {
	accept := r.Header.Get("Accept")
	if accept == "" || accept == "*/*" {
		return true
	}
	return strings.Contains(accept, contentType)
}

// BearerToken extracts the Bearer token from the Authorization header.
// Returns empty string if not present or not a Bearer token.
func BearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

// BasicAuth extracts username and password from Basic auth header.
// Wraps http.Request.BasicAuth for convenience.
func BasicAuth(r *http.Request) (username, password string, ok bool) {
	return r.BasicAuth()
}

// WriteJSON is a shorthand for writing a JSON response with status code.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error response in the format {"error": message}.
func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, map[string]string{"error": message})
}

// FullURL returns the full URL of the request including scheme and host.
func FullURL(r *http.Request) string {
	scheme := "http"
	if IsTLS(r) {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
}
