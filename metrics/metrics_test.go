package metrics

import (
	"math"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
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

	if c.Value() != 3 {
		t.Errorf("expected 3, got %d", c.Value())
	}
}

func TestCounter_Add(t *testing.T) {
	c := NewCounter("bytes_total", "Total bytes")
	c.Add(100)
	c.Add(50)
	c.Add(-10) // Should be ignored.

	if c.Value() != 150 {
		t.Errorf("expected 150, got %d", c.Value())
	}
}

func TestCounter_Name_Help(t *testing.T) {
	c := NewCounter("test_counter", "A counter")
	if c.Name() != "test_counter" {
		t.Error("wrong name")
	}
	if c.Help() != "A counter" {
		t.Error("wrong help")
	}
}

func TestCounter_WithLabels(t *testing.T) {
	c := NewCounter("http_requests_total", "HTTP requests")
	v := c.WithLabels(map[string]string{"method": "GET", "path": "/api"})
	v.Add(5)

	v2 := c.WithLabels(map[string]string{"method": "GET", "path": "/api"})
	if v2.Load() != 5 {
		t.Errorf("expected 5 for same labels, got %d", v2.Load())
	}

	v3 := c.WithLabels(map[string]string{"method": "POST", "path": "/api"})
	v3.Add(1)
	if v3.Load() != 1 {
		t.Errorf("expected 1 for different labels, got %d", v3.Load())
	}
}

func TestCounter_Write(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")
	c.Inc()
	c.Inc()

	output := c.Write()
	if !strings.Contains(output, "# HELP requests_total Total requests") {
		t.Error("missing HELP line")
	}
	if !strings.Contains(output, "# TYPE requests_total counter") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(output, "requests_total 2") {
		t.Errorf("expected value 2 in output: %s", output)
	}
}

func TestCounter_Write_WithLabels(t *testing.T) {
	c := NewCounter("http_requests_total", "HTTP requests")
	c.WithLabels(map[string]string{"method": "GET"}).Add(10)

	output := c.Write()
	if !strings.Contains(output, `method="GET"`) {
		t.Errorf("expected label in output: %s", output)
	}
}

func TestGauge_SetAndValue(t *testing.T) {
	g := NewGauge("temperature", "Current temperature")
	g.Set(42.5)
	if g.Value() != 42.5 {
		t.Errorf("expected 42.5, got %f", g.Value())
	}
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
	if g.Name() != "test_gauge" {
		t.Error("wrong name")
	}
	if g.Help() != "A gauge" {
		t.Error("wrong help")
	}
}

func TestGauge_Write(t *testing.T) {
	g := NewGauge("cpu_usage", "CPU usage")
	g.Set(0.85)

	output := g.Write()
	if !strings.Contains(output, "# TYPE cpu_usage gauge") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(output, "cpu_usage 0.85") {
		t.Errorf("expected value in output: %s", output)
	}
}

func TestHistogram_Observe(t *testing.T) {
	h := NewHistogram("request_duration_seconds", "Request duration", nil)
	h.Observe(0.1)
	h.Observe(0.5)
	h.Observe(1.0)

	if h.Count() != 3 {
		t.Errorf("expected count 3, got %d", h.Count())
	}
	if math.Abs(h.Sum()-1.6) > 0.001 {
		t.Errorf("expected sum 1.6, got %f", h.Sum())
	}
}

func TestHistogram_CustomBuckets(t *testing.T) {
	h := NewHistogram("custom", "Custom", []float64{1, 5, 10})
	h.Observe(3)
	h.Observe(7)
	h.Observe(0.5)

	if h.Count() != 3 {
		t.Errorf("expected 3, got %d", h.Count())
	}
}

func TestHistogram_Name_Help(t *testing.T) {
	h := NewHistogram("test_hist", "A histogram", nil)
	if h.Name() != "test_hist" {
		t.Error("wrong name")
	}
	if h.Help() != "A histogram" {
		t.Error("wrong help")
	}
}

func TestHistogram_Write(t *testing.T) {
	h := NewHistogram("latency", "Latency", []float64{0.1, 0.5, 1.0})
	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(0.8)

	output := h.Write()
	if !strings.Contains(output, "# TYPE latency histogram") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(output, `latency_bucket{le="0.1"} 1`) {
		t.Errorf("wrong bucket count for 0.1: %s", output)
	}
	if !strings.Contains(output, `latency_bucket{le="0.5"} 2`) {
		t.Errorf("wrong bucket count for 0.5: %s", output)
	}
	if !strings.Contains(output, `latency_bucket{le="1"} 3`) {
		t.Errorf("wrong bucket count for 1.0: %s", output)
	}
	if !strings.Contains(output, `latency_bucket{le="+Inf"} 3`) {
		t.Errorf("wrong +Inf bucket: %s", output)
	}
	if !strings.Contains(output, "latency_count 3") {
		t.Errorf("wrong count: %s", output)
	}
}

func TestTimer_ObserveDuration(t *testing.T) {
	h := NewHistogram("timer_test", "Timer", []float64{0.001, 0.01, 0.1, 1.0})
	timer := NewTimer(h)

	d := timer.ObserveDuration()
	if d <= 0 {
		t.Error("expected positive duration")
	}
	if h.Count() != 1 {
		t.Errorf("expected 1 observation, got %d", h.Count())
	}
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
	if !strings.Contains(output, "req_total 1") {
		t.Errorf("missing counter in output: %s", output)
	}
	if !strings.Contains(output, "active 5") {
		t.Errorf("missing gauge in output: %s", output)
	}
}

func TestRegistry_Handler(t *testing.T) {
	r := NewRegistry()
	r.Register(NewCounter("test_total", "Test"))

	handler := r.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", ct)
	}
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

	if c.Value() != 10000 {
		t.Errorf("expected 10000, got %d", c.Value())
	}
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

	if h.Count() != 1000 {
		t.Errorf("expected 1000, got %d", h.Count())
	}
}

func TestLabelsKey(t *testing.T) {
	key := labelsKey(map[string]string{"b": "2", "a": "1"})
	if key != `a="1",b="2"` {
		t.Errorf("expected sorted key, got %q", key)
	}
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
