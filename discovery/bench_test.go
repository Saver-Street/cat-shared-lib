package discovery

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"
)

func newBenchRegistry(b *testing.B, services, instancesPer int) *Registry {
	b.Helper()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := NewRegistry()
	for s := range services {
		for i := range instancesPer {
			_ = r.Register(Instance{
				Service:  fmt.Sprintf("svc-%d", s),
				ID:       fmt.Sprintf("inst-%d-%d", s, i),
				Addr:     fmt.Sprintf("http://10.0.%d.%d:8080", s, i),
				Metadata: map[string]string{"version": "1.0"},
			})
		}
	}
	return r
}

func BenchmarkRegister(b *testing.B) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := NewRegistry()
	for i := 0; b.Loop(); i++ {
		_ = r.Register(Instance{
			Service: "bench-svc",
			ID:      fmt.Sprintf("inst-%d", i),
			Addr:    fmt.Sprintf("http://10.0.0.%d:8080", i%256),
		})
	}
}

func BenchmarkResolve(b *testing.B) {
	r := newBenchRegistry(b, 5, 10)
	b.ResetTimer()
	for b.Loop() {
		r.Resolve("svc-2")
	}
}

func BenchmarkResolveAll(b *testing.B) {
	r := newBenchRegistry(b, 5, 10)
	b.ResetTimer()
	for b.Loop() {
		r.ResolveAll("svc-2")
	}
}

func BenchmarkResolveHealthy(b *testing.B) {
	r := newBenchRegistry(b, 5, 10)
	// Mark half as unhealthy.
	for i := 0; i < 5; i++ {
		_ = r.SetStatus("svc-2", fmt.Sprintf("inst-2-%d", i), StatusUnhealthy)
	}
	b.ResetTimer()
	for b.Loop() {
		r.ResolveHealthy("svc-2")
	}
}

func BenchmarkHeartbeat(b *testing.B) {
	r := newBenchRegistry(b, 1, 10)
	b.ResetTimer()
	for b.Loop() {
		r.Heartbeat("svc-0", "inst-0-0")
	}
}

func BenchmarkMarkStale(b *testing.B) {
	r := newBenchRegistry(b, 5, 10)
	b.ResetTimer()
	for b.Loop() {
		r.MarkStale(1 * time.Hour)
	}
}

func BenchmarkServices(b *testing.B) {
	r := newBenchRegistry(b, 20, 5)
	b.ResetTimer()
	for b.Loop() {
		r.Services()
	}
}

func BenchmarkDeregister(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		r := newBenchRegistry(b, 1, 1)
		r.Deregister("svc-0", "inst-0-0")
	}
}
