// Package shutdown provides graceful shutdown helpers for HTTP servers,
// including OS signal handling and connection draining with configurable timeouts.
package shutdown

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	wg sync.WaitGroup
}

// Add increments the in-flight connection count.
func (d *Drainer) Add() { d.wg.Add(1) }

// Done decrements the in-flight connection count.
func (d *Drainer) Done() { d.wg.Done() }

// Wait blocks until all in-flight connections have completed.
func (d *Drainer) Wait() { d.wg.Wait() }

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
