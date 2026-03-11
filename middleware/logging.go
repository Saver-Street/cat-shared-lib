package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	code int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.code = code
	sw.ResponseWriter.WriteHeader(code)
}

// Logging returns middleware that logs each request using structured slog.
// It records method, path, status, duration, and request ID (if present).
// The optional logger parameter specifies the slog.Logger to use;
// if nil, slog.Default() is used.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}

			next.ServeHTTP(sw, r)

			l := logger
			if l == nil {
				l = slog.Default()
			}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.code),
				slog.Duration("duration", time.Since(start)),
			}

			if reqID := GetRequestID(r); reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			if userID := GetUserID(r); userID != "" {
				attrs = append(attrs, slog.String("user_id", userID))
			}

			if corrID := GetCorrelationID(r); corrID != "" {
				attrs = append(attrs, slog.String("correlation_id", corrID))
			}

			// Convert []slog.Attr to []any for LogAttrs
			l.LogAttrs(r.Context(), slog.LevelInfo, "middleware: request completed", attrs...)
		})
	}
}
