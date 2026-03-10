package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Saver-Street/cat-shared-lib/response"
)

// DefaultTimeout is the default request timeout when none is specified.
const DefaultTimeout = 30 * time.Second

// Timeout returns middleware that enforces a per-request timeout. If the
// handler does not complete within the given duration, the client receives
// a 504 Gateway Timeout response. The context passed to downstream handlers
// will carry the deadline so database queries, HTTP calls, and other
// context-aware operations can respect it.
//
// A zero or negative duration disables the timeout (the handler runs without
// a deadline). This can be useful in tests or for long-running endpoints
// that opt out by wrapping with Timeout(0).
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next // no-op when timeout is disabled
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{ResponseWriter: w}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Handler finished normally.
			case <-ctx.Done():
				// Timeout expired — only write if handler hasn't started writing.
				if tw.tryTimeout() {
					response.Error(w, http.StatusGatewayTimeout, "Request timed out")
				}
				// Wait for handler goroutine to finish to avoid leaking.
				<-done
			}
		})
	}
}

// timeoutWriter tracks whether the handler has begun writing the response.
// All fields are guarded by mu since the handler goroutine and the timeout
// goroutine may access them concurrently.
type timeoutWriter struct {
	http.ResponseWriter
	mu          sync.Mutex
	wroteHeader bool
	timedOut    bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return // suppress writes after timeout
	}
	tw.wroteHeader = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, context.DeadlineExceeded
	}
	tw.wroteHeader = true
	return tw.ResponseWriter.Write(b)
}

// tryTimeout marks the writer as timed out and returns true only if the
// handler has not yet started writing the response.
func (tw *timeoutWriter) tryTimeout() bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.wroteHeader {
		return false
	}
	tw.timedOut = true
	return true
}
