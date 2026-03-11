// Package tracing provides OpenTelemetry distributed tracing setup, span
// helpers, and HTTP middleware for trace context propagation.
//
// Create a [Provider] with [NewProvider] and a [Config] selecting an
// [ExporterType] ([ExporterStdout] or [ExporterNoop]).  Use
// [Provider.Tracer] to obtain a named tracer, and [Start] or
// [StartWithTracer] to begin spans.
//
// [RecordError] and [SetAttributes] annotate the current span.  [TraceID],
// [SpanID], and [IsRecording] extract trace context from a context.Context.
//
// [Middleware] wraps an http.Handler to create a span per request.
// [InjectHTTP] and [ExtractHTTP] propagate trace context across service
// boundaries via HTTP headers.
package tracing
