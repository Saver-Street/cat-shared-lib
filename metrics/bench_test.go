package metrics

import (
	"testing"
)

func BenchmarkCounter_Add(b *testing.B) {
	c := NewCounter("bench_add", "Benchmark counter add")
	for b.Loop() {
		c.Add(1)
	}
}

func BenchmarkCounter_Value(b *testing.B) {
	c := NewCounter("bench_value", "Benchmark counter value")
	c.Add(100)
	for b.Loop() {
		_ = c.Value()
	}
}

func BenchmarkCounter_Write(b *testing.B) {
	c := NewCounter("bench_write", "Benchmark counter write")
	c.Add(42)
	for b.Loop() {
		_ = c.Write()
	}
}

func BenchmarkCounter_WithLabels(b *testing.B) {
	c := NewCounter("bench_labels", "Benchmark counter labels")
	labels := map[string]string{"method": "GET", "status": "200"}
	for b.Loop() {
		c.WithLabels(labels).Add(1)
	}
}

func BenchmarkGauge_Set(b *testing.B) {
	g := NewGauge("bench_gauge_set", "Benchmark gauge set")
	for b.Loop() {
		g.Set(42.0)
	}
}

func BenchmarkGauge_Add(b *testing.B) {
	g := NewGauge("bench_gauge_add", "Benchmark gauge add")
	for b.Loop() {
		g.Add(1.5)
	}
}

func BenchmarkGauge_Write(b *testing.B) {
	g := NewGauge("bench_gauge_write", "Benchmark gauge write")
	g.Set(42.0)
	for b.Loop() {
		_ = g.Write()
	}
}

func BenchmarkRegistry_Expose(b *testing.B) {
	r := NewRegistry()
	c := NewCounter("req_total", "Total requests")
	c.Add(1000)
	g := NewGauge("queue_size", "Queue size")
	g.Set(42)
	h := NewHistogram("duration", "Duration", nil)
	h.Observe(0.5)
	r.Register(c)
	r.Register(g)
	r.Register(h)
	for b.Loop() {
		_ = r.Expose()
	}
}

func BenchmarkNewRegistry(b *testing.B) {
	for b.Loop() {
		_ = NewRegistry()
	}
}
