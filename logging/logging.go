// Package logging provides structured logging helpers built on log/slog.
//
// It offers opinionated constructors for creating pre-configured loggers,
// context-based attribute propagation, and common attribute builders
// for use across services in the cat-shared-lib ecosystem.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

// New creates a new slog.Logger with the given options applied.
// By default, the logger writes JSON to os.Stdout at LevelInfo.
func New(opts ...Option) *slog.Logger {
	cfg := config{
		writer: os.Stdout,
		level:  slog.LevelInfo,
		format: "json",
	}
	for _, o := range opts {
		o(&cfg)
	}

	hopts := &slog.HandlerOptions{
		Level:       cfg.level,
		AddSource:   cfg.addSource,
		ReplaceAttr: cfg.replaceAttr,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.format) {
	case "text":
		handler = slog.NewTextHandler(cfg.writer, hopts)
	default:
		handler = slog.NewJSONHandler(cfg.writer, hopts)
	}

	if len(cfg.attrs) > 0 {
		handler = handler.WithAttrs(cfg.attrs)
	}

	return slog.New(handler)
}

// ParseLevel converts a string to a slog.Level.
// Accepted values (case-insensitive): debug, info, warn, error.
// Unknown strings default to LevelInfo.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns a new context carrying the given logger.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext extracts the logger from context. If none is present,
// slog.Default() is returned.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// With returns a logger derived from the context logger with additional
// attributes. If no logger is in the context, slog.Default() is used.
func With(ctx context.Context, args ...any) *slog.Logger {
	return FromContext(ctx).With(args...)
}

// --- Common attribute helpers ---

// RequestID returns a slog.Attr for a request ID.
func RequestID(id string) slog.Attr { return slog.String("request_id", id) }

// UserID returns a slog.Attr for a user ID.
func UserID(id string) slog.Attr { return slog.String("user_id", id) }

// Service returns a slog.Attr for a service name.
func Service(name string) slog.Attr { return slog.String("service", name) }

// Version returns a slog.Attr for a service version.
func Version(v string) slog.Attr { return slog.String("version", v) }

// Component returns a slog.Attr for a logical component within a service.
func Component(name string) slog.Attr { return slog.String("component", name) }

// TraceID returns a slog.Attr for a distributed trace ID.
func TraceID(id string) slog.Attr { return slog.String("trace_id", id) }

// Err returns a slog.Attr for an error value.
func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("error", err.Error())
}

// --- Options ---

type config struct {
	writer      io.Writer
	level       slog.Level
	format      string
	addSource   bool
	attrs       []slog.Attr
	replaceAttr func(groups []string, a slog.Attr) slog.Attr
}

// Option configures a logger created by New.
type Option func(*config)

// WithWriter sets the output writer (default: os.Stdout).
func WithWriter(w io.Writer) Option { return func(c *config) { c.writer = w } }

// WithLevel sets the minimum log level (default: LevelInfo).
func WithLevel(l slog.Level) Option { return func(c *config) { c.level = l } }

// WithFormat sets the output format: "json" (default) or "text".
func WithFormat(f string) Option { return func(c *config) { c.format = f } }

// WithSource enables source file and line information in log entries.
func WithSource() Option { return func(c *config) { c.addSource = true } }

// WithAttrs sets default attributes attached to every log entry.
func WithAttrs(attrs ...slog.Attr) Option { return func(c *config) { c.attrs = attrs } }

// WithReplaceAttr sets a function to transform attributes before they are logged.
func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(c *config) { c.replaceAttr = fn }
}
