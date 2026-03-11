// Package metrics provides Prometheus-compatible instrumentation primitives
// (counters, gauges, histograms) with a registry that exposes metrics in
// the Prometheus text exposition format.
package metrics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Registry holds all registered metrics and exposes them via HTTP.
type Registry struct {
	mu      sync.RWMutex
	metrics map[string]Metric
	order   []string
}

// Metric is the interface all metric types implement.
type Metric interface {
	Name() string
	Help() string
	Write() string
}

// Compile-time interface compliance checks.
var (
	_ Metric = (*Counter)(nil)
	_ Metric = (*Gauge)(nil)
	_ Metric = (*Histogram)(nil)
)

// NewRegistry creates a new metric registry.
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

// Register adds a metric to the registry. Panics on duplicate names.
func (r *Registry) Register(m Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.metrics[m.Name()]; exists {
		panic(fmt.Sprintf("metrics: duplicate metric name %q", m.Name()))
	}
	r.metrics[m.Name()] = m
	r.order = append(r.order, m.Name())
}

// Handler returns an HTTP handler that serves metrics in Prometheus text format.
func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Expose()))
	}
}

// Expose returns all metrics in Prometheus text exposition format.
func (r *Registry) Expose() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder
	for _, name := range r.order {
		m := r.metrics[name]
		sb.WriteString(m.Write())
	}
	return sb.String()
}

// --- Counter ---

// Counter is a monotonically increasing counter.
type Counter struct {
	name   string
	help   string
	value  atomic.Int64
	labels map[string]*atomic.Int64
	mu     sync.RWMutex
}

// NewCounter creates a new counter metric.
func NewCounter(name, help string) *Counter {
	return &Counter{
		name:   name,
		help:   help,
		labels: make(map[string]*atomic.Int64),
	}
}

func (c *Counter) Name() string { return c.name }
func (c *Counter) Help() string { return c.help }

// Inc increments the counter by 1.
func (c *Counter) Inc() { c.value.Add(1) }

// Add increments the counter by the given positive value.
func (c *Counter) Add(v int64) {
	if v > 0 {
		c.value.Add(v)
	}
}

// Value returns the current counter value.
func (c *Counter) Value() int64 { return c.value.Load() }

// WithLabels returns a counter scoped to the given label set.
func (c *Counter) WithLabels(labels map[string]string) *atomic.Int64 {
	key := labelsKey(labels)
	c.mu.RLock()
	if v, ok := c.labels[key]; ok {
		c.mu.RUnlock()
		return v
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.labels[key]; ok {
		return v
	}
	v := &atomic.Int64{}
	c.labels[key] = v
	return v
}

func (c *Counter) Write() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# HELP %s %s\n", c.name, c.help)
	fmt.Fprintf(&sb, "# TYPE %s counter\n", c.name)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.labels) == 0 {
		fmt.Fprintf(&sb, "%s %d\n", c.name, c.value.Load())
	} else {
		keys := sortedKeys(c.labels)
		for _, k := range keys {
			fmt.Fprintf(&sb, "%s{%s} %d\n", c.name, k, c.labels[k].Load())
		}
	}
	return sb.String()
}

// --- Gauge ---

// Gauge is a metric that can go up and down.
type Gauge struct {
	name string
	help string
	bits atomic.Uint64
}

// NewGauge creates a new gauge metric.
func NewGauge(name, help string) *Gauge {
	g := &Gauge{name: name, help: help}
	g.bits.Store(math.Float64bits(0))
	return g
}

func (g *Gauge) Name() string { return g.name }
func (g *Gauge) Help() string { return g.help }

// Set sets the gauge to the given value.
func (g *Gauge) Set(v float64) {
	g.bits.Store(math.Float64bits(v))
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() { g.Add(1) }

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() { g.Add(-1) }

// Add adds the given value to the gauge.
func (g *Gauge) Add(delta float64) {
	for {
		old := g.bits.Load()
		newVal := math.Float64frombits(old) + delta
		if g.bits.CompareAndSwap(old, math.Float64bits(newVal)) {
			return
		}
	}
}

// Value returns the current gauge value.
func (g *Gauge) Value() float64 {
	return math.Float64frombits(g.bits.Load())
}

func (g *Gauge) Write() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# HELP %s %s\n", g.name, g.help)
	fmt.Fprintf(&sb, "# TYPE %s gauge\n", g.name)
	fmt.Fprintf(&sb, "%s %g\n", g.name, g.Value())
	return sb.String()
}

// --- Histogram ---

// Histogram tracks the distribution of observed values in configurable buckets.
type Histogram struct {
	name    string
	help    string
	buckets []float64
	counts  []atomic.Uint64
	sum     atomic.Uint64
	count   atomic.Uint64
}

// DefaultBuckets are the default histogram bucket boundaries.
var DefaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// NewHistogram creates a new histogram. If buckets is nil, DefaultBuckets is used.
func NewHistogram(name, help string, buckets []float64) *Histogram {
	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}
	sorted := make([]float64, len(buckets))
	copy(sorted, buckets)
	sort.Float64s(sorted)

	return &Histogram{
		name:    name,
		help:    help,
		buckets: sorted,
		counts:  make([]atomic.Uint64, len(sorted)),
	}
}

func (h *Histogram) Name() string { return h.name }
func (h *Histogram) Help() string { return h.help }

// Observe records a value in the histogram.
func (h *Histogram) Observe(v float64) {
	for i, bound := range h.buckets {
		if v <= bound {
			h.counts[i].Add(1)
			break
		}
	}
	// Atomically add to sum using CAS.
	for {
		old := h.sum.Load()
		newVal := math.Float64frombits(old) + v
		if h.sum.CompareAndSwap(old, math.Float64bits(newVal)) {
			break
		}
	}
	h.count.Add(1)
}

// Sum returns the sum of observed values.
func (h *Histogram) Sum() float64 {
	return math.Float64frombits(h.sum.Load())
}

// Count returns the total number of observations.
func (h *Histogram) Count() uint64 {
	return h.count.Load()
}

func (h *Histogram) Write() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# HELP %s %s\n", h.name, h.help)
	fmt.Fprintf(&sb, "# TYPE %s histogram\n", h.name)

	var cumulative uint64
	for i, bound := range h.buckets {
		cumulative += h.counts[i].Load()
		fmt.Fprintf(&sb, "%s_bucket{le=\"%g\"} %d\n", h.name, bound, cumulative)
	}
	fmt.Fprintf(&sb, "%s_bucket{le=\"+Inf\"} %d\n", h.name, h.count.Load())
	fmt.Fprintf(&sb, "%s_sum %g\n", h.name, h.Sum())
	fmt.Fprintf(&sb, "%s_count %d\n", h.name, h.count.Load())
	return sb.String()
}

// --- Timer ---

// Timer is a helper that measures duration and observes it in a histogram.
type Timer struct {
	histogram *Histogram
	start     time.Time
}

// NewTimer starts a new timer that will observe into the given histogram.
func NewTimer(h *Histogram) *Timer {
	return &Timer{histogram: h, start: time.Now()}
}

// ObserveDuration stops the timer and records the elapsed duration in seconds.
func (t *Timer) ObserveDuration() time.Duration {
	d := time.Since(t.start)
	t.histogram.Observe(d.Seconds())
	return d
}

// --- Helpers ---

func labelsKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", k, labels[k]))
	}
	return strings.Join(parts, ",")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
