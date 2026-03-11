// Package logging provides structured logging helpers built on log/slog.
//
// It offers opinionated constructors for creating pre-configured loggers,
// context-based attribute propagation, and common attribute builders
// for use across services in the cat-shared-lib ecosystem.
//
// # Creating a Logger
//
// Use [New] with functional options to create a configured logger:
//
//	logger := logging.New(
//	    logging.WithFormat("json"),
//	    logging.WithLevel(slog.LevelInfo),
//	    logging.WithAttrs(logging.Service("api"), logging.Version("1.0")),
//	)
//
// # Context Propagation
//
// Store a logger in context with [WithContext] and retrieve it with
// [FromContext]. Use [With] to add attributes to the context logger:
//
//	ctx = logging.WithContext(ctx, logger)
//	logging.FromContext(ctx).Info("request handled")
//
// # Attribute Helpers
//
// Common attribute builders like [RequestID], [UserID], [Service], [TraceID],
// and [Err] produce slog.Attr values with standardised key names for
// consistent structured logs across all services.
package logging
