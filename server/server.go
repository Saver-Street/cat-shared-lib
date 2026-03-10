// Package server provides a standard HTTP server with graceful shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Config controls server behavior.
type Config struct {
	Addr            string
	Handler         http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// defaults fills in zero-value fields with sensible defaults.
func (c *Config) defaults() {
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 15 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 60 * time.Second
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 10 * time.Second
	}
}

// ListenAndServe starts the HTTP server and blocks until a SIGINT or SIGTERM
// is received, then shuts down gracefully within the configured timeout.
// cleanup functions run after server shutdown (e.g., closing DB pools).
// Returns an error immediately if cfg.Handler is nil.
func ListenAndServe(cfg Config, cleanup ...func()) error {
	if cfg.Handler == nil {
		return fmt.Errorf("server: Handler must not be nil")
	}
	cfg.defaults()

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      cfg.Handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	runCleanup := func() {
		for _, fn := range cleanup {
			fn()
		}
	}

	select {
	case err := <-errCh:
		runCleanup()
		return err
	case sig := <-done:
		slog.Info("received signal, shutting down", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		runCleanup()
		return err
	}

	runCleanup()
	slog.Info("server stopped")
	return nil
}
