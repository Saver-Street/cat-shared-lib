// Package metrics provides Prometheus-compatible instrumentation primitives
// and an HTTP handler that exposes metrics in the text exposition format.
//
// Create a [Registry] with [NewRegistry] and register [Counter], [Gauge], and
// [Histogram] instances via [Registry.Register].  Use [Registry.Handler] or
// [Registry.Expose] to serve collected metrics at an HTTP endpoint.
//
// [Counter] tracks monotonically increasing values; [Gauge] holds arbitrary
// numeric values that can go up and down; [Histogram] records observations
// into configurable buckets (see [DefaultBuckets]).  [Timer] is a convenience
// wrapper that measures a duration and records it to a [Histogram] via
// [Timer.ObserveDuration].
//
// Counters and histograms support [Counter.WithLabels] for dimensional
// metrics.
package metrics
