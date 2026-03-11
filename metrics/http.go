package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// HTTPMiddleware returns HTTP middleware that records per-request metrics.
// It tracks request count (by method and status) and request duration.
// The metrics are registered with the provided registry.
func HTTPMiddleware(reg *Registry) func(http.Handler) http.Handler {
	reqCount := NewCounter("http_requests_total", "Total number of HTTP requests")
	reqDuration := NewHistogram("http_request_duration_seconds", "HTTP request duration in seconds", DefaultBuckets)
	reg.Register(reqCount)
	reg.Register(reqDuration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusCapturer{ResponseWriter: w, code: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start).Seconds()
			labels := map[string]string{
				"method": r.Method,
				"status": strconv.Itoa(sw.code),
			}
			reqCount.WithLabels(labels).Add(1)
			reqDuration.Observe(duration)
		})
	}
}

type statusCapturer struct {
	http.ResponseWriter
	code    int
	written bool
}

func (s *statusCapturer) WriteHeader(code int) {
	if !s.written {
		s.code = code
		s.written = true
	}
	s.ResponseWriter.WriteHeader(code)
}
