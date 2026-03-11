package metrics

import (
	"math"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	testkit.RequireNotNil(t, r)
}

func TestRegistry_Register_Duplicate_Panics(t *testing.T) {
	r := NewRegistry()
	c := NewCounter("test_total", "test help")
	r.Register(c)

	testkit.AssertPanics(t, func() {
		r.Register(NewCounter("test_total", "duplicate"))
	})
}

func TestCounter_IncAndValue(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")
	c.Inc()
	c.Inc()
	c.Inc()

	testkit.AssertEqual(t, c.Value(), int64(3))
}

func TestCounter_Add(t *testing.T) {
	c := NewCounter("bytes_total", "Total bytes")
	c.Add(100)
	c.Add(50)
	c.Add(-10) // Should be ignored.

	testkit.AssertEqual(t, c.Value(), int64(150))
}

func TestCounter_Name_Help(t *testing.T) {
	c := NewCounter("test_counter", "A counter")
	testkit.AssertEqual(t, c.Name(), "test_counter")
	testkit.AssertEqual(t, c.Help(), "A counter")
}

func TestCounter_WithLabels(t *testing.T) {
	c := NewCounter("http_requests_total", "HTTP requests")
	v := c.WithLabels(map[string]string{"method": "GET", "path": "/api"})
	v.Add(5)

	v2 := c.WithLabels(map[string]string{"method": "GET", "path": "/api"})
	testkit.AssertEqual(t, v2.Load(), int64(5))

	v3 := c.WithLabels(map[string]string{"method": "POST", "path": "/api"})
	v3.Add(1)
	testkit.AssertEqual(t, v3.Load(), int64(1))
}

func TestCounter_Write(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")
	c.Inc()
	c.Inc()

	output := c.Write()
	testkit.AssertContains(t, output, "# HELP requests_total Total requests")
	testkit.AssertContains(t, output, "# TYPE requests_total counter")
	testkit.AssertContains(t, output, "requests_total 2")
}

func TestCounter_Write_WithLabels(t *testing.T) {
	c := NewCounter("http_requests_total", "HTTP requests")
	c.WithLabels(map[string]string{"method": "GET"}).Add(10)

	output := c.Write()
	testkit.AssertContains(t, output, `method="GET"`)
}

func TestGauge_SetAndValue(t *testing.T) {
	g := NewGauge("temperature", "Current temperature")
	g.Set(42.5)
	testkit.AssertEqual(t, g.Value(), 42.5)
}

func TestGauge_IncDec(t *testing.T) {
	g := NewGauge("connections", "Active connections")
	g.Inc()
	g.Inc()
	g.Dec()

	if math.Abs(g.Value()-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", g.Value())
	}
}

func TestGauge_Add(t *testing.T) {
	g := NewGauge("queue_size", "Queue size")
	g.Add(5)
	g.Add(-2)

	if math.Abs(g.Value()-3.0) > 0.001 {
		t.Errorf("expected 3.0, got %f", g.Value())
	}
}

func TestGauge_Name_Help(t *testing.T) {
	g := NewGauge("test_gauge", "A gauge")
	testkit.AssertEqual(t, g.Name(), "test_gauge")
	testkit.AssertEqual(t, g.Help(), "A gauge")
}

func TestGauge_Write(t *testing.T) {
	g := NewGauge("cpu_usage", "CPU usage")
	g.Set(0.85)

	output := g.Write()
	testkit.AssertContains(t, output, "# TYPE cpu_usage gauge")
	testkit.AssertContains(t, output, "cpu_usage 0.85")
}

func TestHistogram_Observe(t *testing.T) {
	h := NewHistogram("request_duration_seconds", "Request duration", nil)
	h.Observe(0.1)
	h.Observe(0.5)
	h.Observe(1.0)

	testkit.AssertEqual(t, h.Count(), uint64(3))
	if math.Abs(h.Sum()-1.6) > 0.001 {
		t.Errorf("expected sum 1.6, got %f", h.Sum())
	}
}

func TestHistogram_CustomBuckets(t *testing.T) {
	h := NewHistogram("custom", "Custom", []float64{1, 5, 10})
	h.Observe(3)
	h.Observe(7)
	h.Observe(0.5)

	testkit.AssertEqual(t, h.Count(), uint64(3))
}

func TestHistogram_Name_Help(t *testing.T) {
	h := NewHistogram("test_hist", "A histogram", nil)
	testkit.AssertEqual(t, h.Name(), "test_hist")
	testkit.AssertEqual(t, h.Help(), "A histogram")
}

func TestHistogram_Write(t *testing.T) {
	h := NewHistogram("latency", "Latency", []float64{0.1, 0.5, 1.0})
	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(0.8)

	output := h.Write()
	testkit.AssertContains(t, output, "# TYPE latency histogram")
	testkit.AssertContains(t, output, `latency_bucket{le="0.1"} 1`)
	testkit.AssertContains(t, output, `latency_bucket{le="0.5"} 2`)
	testkit.AssertContains(t, output, `latency_bucket{le="1"} 3`)
	testkit.AssertContains(t, output, `latency_bucket{le="+Inf"} 3`)
	testkit.AssertContains(t, output, "latency_count 3")
}

func TestTimer_ObserveDuration(t *testing.T) {
	h := NewHistogram("timer_test", "Timer", []float64{0.001, 0.01, 0.1, 1.0})
	timer := NewTimer(h)

	d := timer.ObserveDuration()
	if d <= 0 {
		t.Error("expected positive duration")
	}
	testkit.AssertEqual(t, h.Count(), uint64(1))
}

func TestRegistry_Expose(t *testing.T) {
	r := NewRegistry()
	c := NewCounter("req_total", "Total requests")
	g := NewGauge("active", "Active conns")
	c.Inc()
	g.Set(5)
	r.Register(c)
	r.Register(g)

	output := r.Expose()
	testkit.AssertContains(t, output, "req_total 1")
	testkit.AssertContains(t, output, "active 5")
}

func TestRegistry_Handler(t *testing.T) {
	r := NewRegistry()
	r.Register(NewCounter("test_total", "Test"))

	handler := r.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, rr.Code, 200)
	ct := rr.Header().Get("Content-Type")
	testkit.AssertContains(t, ct, "text/plain")
}

func TestConcurrent_Counter(t *testing.T) {
	c := NewCounter("concurrent", "Test")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	testkit.AssertEqual(t, c.Value(), int64(10000))
}

func TestConcurrent_Gauge(t *testing.T) {
	g := NewGauge("concurrent_gauge", "Test")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Inc()
			g.Dec()
		}()
	}
	wg.Wait()
	// After equal incs and decs, should be ~0.
	if math.Abs(g.Value()) > 0.001 {
		t.Errorf("expected ~0, got %f", g.Value())
	}
}

func TestConcurrent_Histogram(t *testing.T) {
	h := NewHistogram("concurrent_hist", "Test", nil)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				h.Observe(0.1)
			}
		}()
	}
	wg.Wait()

	testkit.AssertEqual(t, h.Count(), uint64(1000))
}

func TestLabelsKey(t *testing.T) {
	key := labelsKey(map[string]string{"b": "2", "a": "1"})
	testkit.AssertEqual(t, key, `a="1",b="2"`)
}

func BenchmarkCounter_Inc(b *testing.B) {
	c := NewCounter("bench", "bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkHistogram_Observe(b *testing.B) {
	h := NewHistogram("bench", "bench", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Observe(0.1)
	}
}

func TestCounter_WithLabels_ConcurrentFirstAccess(t *testing.T) {
	labels := map[string]string{"method": "GET", "path": "/api"}

	// Run many iterations to reliably hit the double-checked locking path:
	// goroutine A creates the key while goroutine B waits for the write lock,
	// then B finds the key already present at line 119.
	for i := 0; i < 200; i++ {
		c := NewCounter("race_test", "test")
		ready := make(chan struct{})
		var wg sync.WaitGroup

		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				c.WithLabels(labels).Add(1)
			}()
		}

		close(ready) // release all goroutines simultaneously
		wg.Wait()
	}
}
