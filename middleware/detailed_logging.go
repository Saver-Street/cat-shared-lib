package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseCapture wraps http.ResponseWriter to capture status code and response size.
type responseCapture struct {
	http.ResponseWriter
	code int
	size int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.code = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	n, err := rc.ResponseWriter.Write(b)
	rc.size += n
	return n, err
}

// DetailedLogging returns middleware that logs comprehensive request and response
// information including method, path, query string, status, duration, response size,
// remote address, user agent, and content type. It also includes request_id and
// user_id if present in context.
//
// If logger is nil, slog.Default() is used.
func DetailedLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rc := &responseCapture{ResponseWriter: w, code: http.StatusOK}

			next.ServeHTTP(rc, r)

			l := logger
			if l == nil {
				l = slog.Default()
			}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rc.code),
				slog.Duration("duration", time.Since(start)),
				slog.Int("response_bytes", rc.size),
				slog.String("remote_addr", r.RemoteAddr),
			}

			if q := r.URL.RawQuery; q != "" {
				attrs = append(attrs, slog.String("query", q))
			}

			if ua := r.UserAgent(); ua != "" {
				attrs = append(attrs, slog.String("user_agent", ua))
			}

			if ct := r.Header.Get("Content-Type"); ct != "" {
				attrs = append(attrs, slog.String("content_type", ct))
			}

			if reqID := GetRequestID(r); reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			if userID := GetUserID(r); userID != "" {
				attrs = append(attrs, slog.String("user_id", userID))
			}

			level := slog.LevelInfo
			if rc.code >= 500 {
				level = slog.LevelError
			} else if rc.code >= 400 {
				level = slog.LevelWarn
			}

			l.LogAttrs(r.Context(), level, "middleware: request completed", attrs...)
		})
	}
}
