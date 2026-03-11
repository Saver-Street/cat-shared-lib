// Package shutdown provides graceful shutdown helpers for HTTP servers,
// including OS signal handling and connection draining with configurable timeouts.
package shutdown

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Config configures graceful shutdown behaviour.
type Config struct {
	// Timeout is the maximum time to wait for in-flight requests to complete.
	// Default: 30 seconds.
	Timeout time.Duration
	// Signals is the list of OS signals that trigger shutdown.
	// Default: [SIGINT, SIGTERM].
	Signals []os.Signal
	// Logger is used for shutdown lifecycle messages. Default: slog.Default().
	Logger *slog.Logger
	// OnShutdown is an optional slice of functions to call during shutdown
	// (e.g., closing database pools, flushing buffers). They run in order
	// before the HTTP server is stopped.
	OnShutdown []func(ctx context.Context) error
}

func (c *Config) defaults() {
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	if len(c.Signals) == 0 {
		c.Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}

// Drainer tracks in-flight connections and provides a mechanism to wait
// for them to complete.
type Drainer struct {
	wg        sync.WaitGroup
	active    atomic.Int64
	completed atomic.Int64
}

// Add increments the in-flight connection count.
func (d *Drainer) Add() {
	d.active.Add(1)
	d.wg.Add(1)
}

// Done decrements the in-flight connection count.
func (d *Drainer) Done() {
	d.active.Add(-1)
	d.completed.Add(1)
	d.wg.Done()
}

// Wait blocks until all in-flight connections have completed.
func (d *Drainer) Wait() { d.wg.Wait() }

// Active returns the number of currently in-flight connections.
func (d *Drainer) Active() int64 { return d.active.Load() }

// Completed returns the total number of connections that have completed
// since the Drainer was created.
func (d *Drainer) Completed() int64 { return d.completed.Load() }

// WaitWithContext blocks until all in-flight connections complete or the
// context is cancelled. Returns the context error if cancelled before
// draining finishes.
func (d *Drainer) WaitWithContext(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Middleware returns HTTP middleware that tracks in-flight requests via the Drainer.
func (d *Drainer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.Add()
		defer d.Done()
		next.ServeHTTP(w, r)
	})
}

// ListenAndServe starts the HTTP server and blocks until a shutdown signal is
// received, then gracefully shuts down the server. It returns any error from
// ListenAndServe that is not http.ErrServerClosed.
func ListenAndServe(server *http.Server, cfg Config) error {
	cfg.defaults()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, cfg.Signals...)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	cfg.Logger.Info("server started", "addr", server.Addr)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		cfg.Logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Run OnShutdown hooks.
	for _, fn := range cfg.OnShutdown {
		if err := fn(ctx); err != nil {
			cfg.Logger.Error("shutdown hook error", "error", err)
		}
	}

	cfg.Logger.Info("shutting down server", "timeout", cfg.Timeout.String())
	if err := server.Shutdown(ctx); err != nil {
		cfg.Logger.Error("server shutdown error", "error", err)
		return err
	}

	cfg.Logger.Info("server stopped gracefully")
	return nil
}

// WaitForSignal blocks until one of the configured signals is received and
// returns a context that will be cancelled after the configured timeout.
// Useful when you need signal handling without an HTTP server.
func WaitForSignal(cfg Config) (context.Context, context.CancelFunc) {
	cfg.defaults()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, cfg.Signals...)

	<-sigCh
	signal.Stop(sigCh)

	cfg.Logger.Info("shutdown signal received, starting drain")
	return context.WithTimeout(context.Background(), cfg.Timeout)
}

// Hook is a named shutdown callback. The Name field is used in log messages
// so operators can tell which cleanup step failed or was slow.
type Hook struct {
	Name string
	Fn   func(ctx context.Context) error
}

// AddHook appends a named shutdown hook and returns the Config for chaining.
func (c *Config) AddHook(name string, fn func(ctx context.Context) error) *Config {
	c.OnShutdown = append(c.OnShutdown, func(ctx context.Context) error {
		return fn(ctx)
	})
	c.hooks = append(c.hooks, Hook{Name: name, Fn: fn})
	return c
}

// RunHooks executes the named hooks sequentially, logging each by name.
// It returns the first error encountered (all hooks still run).
func RunHooks(ctx context.Context, logger *slog.Logger, hooks []Hook) error {
	if logger == nil {
		logger = slog.Default()
	}
	var firstErr error
	for _, h := range hooks {
		logger.Info("running shutdown hook", "hook", h.Name)
		if err := h.Fn(ctx); err != nil {
			logger.Error("shutdown hook failed", "hook", h.Name, "error", err)
			if firstErr == nil {
				firstErr = fmt.Errorf("hook %q: %w", h.Name, err)
			}
		} else {
			logger.Info("shutdown hook completed", "hook", h.Name)
		}
	}
	return firstErr
}

// RunHooksParallel executes all hooks concurrently and waits for them to
// finish. It returns a combined error if any hooks fail.
func RunHooksParallel(ctx context.Context, logger *slog.Logger, hooks []Hook) error {
	if logger == nil {
		logger = slog.Default()
	}
	var (
		mu     sync.Mutex
		errs   []error
		wg     sync.WaitGroup
	)
	for _, h := range hooks {
		wg.Add(1)
		go func(hook Hook) {
			defer wg.Done()
			logger.Info("running shutdown hook", "hook", hook.Name)
			if err := hook.Fn(ctx); err != nil {
				logger.Error("shutdown hook failed", "hook", hook.Name, "error", err)
				mu.Lock()
				errs = append(errs, fmt.Errorf("hook %q: %w", hook.Name, err))
				mu.Unlock()
			} else {
				logger.Info("shutdown hook completed", "hook", hook.Name)
			}
		}(h)
	}
	wg.Wait()
	return errors.Join(errs...)
}
