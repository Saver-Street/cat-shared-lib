package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_NoCheckers(t *testing.T) {
	h := Handler("test-service", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var s Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s.Status != "ok" {
		t.Errorf("expected status ok, got %s", s.Status)
	}
	if s.Service != "test-service" {
		t.Errorf("expected service test-service, got %s", s.Service)
	}
	if s.Version != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %s", s.Version)
	}
	if s.Checks != nil {
		t.Errorf("expected nil checks, got %v", s.Checks)
	}
}

func TestHandler_AllHealthy(t *testing.T) {
	db := NewChecker("db", func(ctx context.Context) error { return nil })
	cache := NewChecker("cache", func(ctx context.Context) error { return nil })

	h := Handler("svc", "v2.0.0", db, cache)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Status != "ok" {
		t.Errorf("expected ok, got %s", s.Status)
	}
	if s.Checks["db"] != "ok" {
		t.Errorf("expected db ok, got %s", s.Checks["db"])
	}
	if s.Checks["cache"] != "ok" {
		t.Errorf("expected cache ok, got %s", s.Checks["cache"])
	}
}

func TestHandler_Degraded(t *testing.T) {
	healthy := NewChecker("db", func(ctx context.Context) error { return nil })
	unhealthy := NewChecker("cache", func(ctx context.Context) error {
		return errors.New("connection refused")
	})

	h := Handler("svc", "v1.0.0", healthy, unhealthy)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Status != "degraded" {
		t.Errorf("expected degraded, got %s", s.Status)
	}
	if s.Checks["db"] != "ok" {
		t.Errorf("expected db ok, got %s", s.Checks["db"])
	}
	if s.Checks["cache"] != "connection refused" {
		t.Errorf("expected cache error, got %s", s.Checks["cache"])
	}
}

func TestHandler_ContentType(t *testing.T) {
	h := Handler("svc", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestNewChecker(t *testing.T) {
	c := NewChecker("test", func(ctx context.Context) error { return nil })
	if c.Name() != "test" {
		t.Errorf("expected name test, got %s", c.Name())
	}
	if err := c.Check(context.Background()); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNewChecker_Error(t *testing.T) {
	c := NewChecker("fail", func(ctx context.Context) error {
		return errors.New("down")
	})
	if err := c.Check(context.Background()); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestHandler_SlowChecker(t *testing.T) {
	slow := NewChecker("slow", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			return nil
		}
	})

	h := Handler("svc", "v1.0.0", slow)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandler_MultipleUnhealthy(t *testing.T) {
	c1 := NewChecker("db", func(ctx context.Context) error {
		return errors.New("timeout")
	})
	c2 := NewChecker("redis", func(ctx context.Context) error {
		return errors.New("refused")
	})

	h := Handler("svc", "v1.0.0", c1, c2)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Checks["db"] != "timeout" {
		t.Errorf("expected db timeout, got %s", s.Checks["db"])
	}
	if s.Checks["redis"] != "refused" {
		t.Errorf("expected redis refused, got %s", s.Checks["redis"])
	}
}

func TestHandler_ContextCancelled(t *testing.T) {
	checker := NewChecker("slow", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	h := Handler("svc", "v1.0.0", checker)
	rr := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	req := httptest.NewRequest("GET", "/health", nil).WithContext(ctx)
	h.ServeHTTP(rr, req)

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Status != "degraded" {
		t.Errorf("cancelled context should cause degraded, got %s", s.Status)
	}
}

func TestHandler_SingleChecker(t *testing.T) {
	c := NewChecker("single", func(ctx context.Context) error { return nil })
	h := Handler("svc", "v1.0.0", c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	if s.Checks["single"] != "ok" {
		t.Errorf("expected single=ok, got %s", s.Checks["single"])
	}
}

func BenchmarkHandler_NoCheckers(b *testing.B) {
	h := Handler("svc", "v1.0.0")
	req := httptest.NewRequest("GET", "/health", nil)
	for b.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkHandler_WithCheckers(b *testing.B) {
	c1 := NewChecker("db", func(ctx context.Context) error { return nil })
	c2 := NewChecker("cache", func(ctx context.Context) error { return nil })
	h := Handler("svc", "v1.0.0", c1, c2)
	req := httptest.NewRequest("GET", "/health", nil)
	for b.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkNewChecker(b *testing.B) {
	fn := func(ctx context.Context) error { return nil }
	for b.Loop() {
		NewChecker("test", fn)
	}
}

// errWriter is a ResponseWriter whose Write method always returns an error.
type errWriter struct {
	header http.Header
	code   int
}

func (e *errWriter) Header() http.Header {
	if e.header == nil {
		e.header = make(http.Header)
	}
	return e.header
}
func (e *errWriter) WriteHeader(code int) { e.code = code }
func (e *errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write: broken pipe")
}

func TestHandler_EncodingError(t *testing.T) {
	h := Handler("svc", "v1.0.0")
	req := httptest.NewRequest("GET", "/health", nil)
	w := &errWriter{}
	// Must not panic when encoding fails.
	h.ServeHTTP(w, req)
}
