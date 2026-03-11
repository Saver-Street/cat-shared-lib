package metrics

import (
	"math"
	"testing"
)

func FuzzCounterAdd(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(0))
	f.Add(int64(-1))
	f.Add(int64(math.MaxInt64))
	f.Fuzz(func(t *testing.T, v int64) {
		c := NewCounter("fuzz_counter", "test")
		c.Add(v)
		// Negative values should be ignored.
		if v <= 0 && c.Value() != 0 {
			t.Errorf("Value = %d after Add(%d), want 0", c.Value(), v)
		}
		if v > 0 && c.Value() != v {
			t.Errorf("Value = %d after Add(%d), want %d", c.Value(), v, v)
		}
		// Write must not panic.
		_ = c.Write()
	})
}

func FuzzCounterWithLabels(f *testing.F) {
	f.Add("method", "GET")
	f.Add("", "")
	f.Add("status", "200")
	f.Add("path", "/api/v1/users")
	f.Fuzz(func(t *testing.T, key, value string) {
		c := NewCounter("fuzz_labeled", "test")
		labels := map[string]string{key: value}
		v := c.WithLabels(labels)
		v.Add(1)
		// Write must not panic and must include label output.
		out := c.Write()
		if out == "" {
			t.Error("Write returned empty")
		}
	})
}

func FuzzGaugeAdd(f *testing.F) {
	f.Add(1.0)
	f.Add(-1.0)
	f.Add(0.0)
	f.Add(math.MaxFloat64)
	f.Add(math.SmallestNonzeroFloat64)
	f.Add(math.NaN())
	f.Add(math.Inf(1))
	f.Add(math.Inf(-1))
	f.Fuzz(func(t *testing.T, v float64) {
		g := NewGauge("fuzz_gauge", "test")
		// Must not panic on any float value.
		g.Set(v)
		g.Add(v)
		g.Inc()
		g.Dec()
		_ = g.Value()
		_ = g.Write()
	})
}

func FuzzHistogramObserve(f *testing.F) {
	f.Add(0.001)
	f.Add(0.5)
	f.Add(1.0)
	f.Add(100.0)
	f.Add(-1.0)
	f.Add(0.0)
	f.Add(math.Inf(1))
	f.Add(math.NaN())
	f.Fuzz(func(t *testing.T, v float64) {
		h := NewHistogram("fuzz_hist", "test", nil)
		// Must not panic on any float value.
		h.Observe(v)
		_ = h.Sum()
		_ = h.Count()
		_ = h.Write()
	})
}

func FuzzRegistryExpose(f *testing.F) {
	f.Add("metric_name", "A help string")
	f.Add("", "")
	f.Add("unicode_名前", "ヘルプ")
	f.Add("name with spaces", "help\nwith\nnewlines")
	f.Fuzz(func(t *testing.T, name, help string) {
		r := NewRegistry()
		c := NewCounter(name, help)
		c.Inc()
		r.Register(c)
		// Expose must not panic.
		out := r.Expose()
		if out == "" {
			t.Error("Expose returned empty")
		}
	})
}
