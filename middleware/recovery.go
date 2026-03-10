package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// Recovery is middleware that recovers from panics in downstream handlers,
// logs the stack trace, and returns a 500 Internal Server Error response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())

				attrs := []slog.Attr{
					slog.Any("panic", rec),
					slog.String("stack", stack),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				}
				if id := GetRequestID(r); id != "" {
					attrs = append(attrs, slog.String("request_id", id))
				}

				slog.LogAttrs(r.Context(), slog.LevelError, "middleware: panic recovered", attrs...)
				response.Error(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
