// Package health provides standardized health check handlers for microservices.
package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Status represents a health check response.
type Status struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// Checker performs a named health check and returns an error if unhealthy.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// checkerFunc adapts a function to the Checker interface.
type checkerFunc struct {
	name    string
	checkFn func(ctx context.Context) error
}

func (c *checkerFunc) Name() string                    { return c.name }
func (c *checkerFunc) Check(ctx context.Context) error { return c.checkFn(ctx) }

// NewChecker creates a Checker from a name and function.
func NewChecker(name string, fn func(ctx context.Context) error) Checker {
	return &checkerFunc{name: name, checkFn: fn}
}

// Handler returns an http.HandlerFunc that responds with health status JSON.
// Checkers are run concurrently with a 5-second timeout.
func Handler(service, version string, checkers ...Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := Status{
			Status:  "ok",
			Service: service,
			Version: version,
		}

		if len(checkers) > 0 {
			status.Checks = make(map[string]string, len(checkers))
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			var mu sync.Mutex
			var wg sync.WaitGroup
			for _, c := range checkers {
				wg.Add(1)
				go func(c Checker) {
					defer wg.Done()
					err := c.Check(ctx)
					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						status.Checks[c.Name()] = err.Error()
						status.Status = "degraded"
					} else {
						status.Checks[c.Name()] = "ok"
					}
				}(c)
			}
			wg.Wait()
		}

		code := http.StatusOK
		if status.Status != "ok" {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(status); err != nil {
			slog.Error("health: failed to encode response", "error", err)
		}
	}
}
