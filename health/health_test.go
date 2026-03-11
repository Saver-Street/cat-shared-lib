package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHandler_NoCheckers(t *testing.T) {
	h := Handler("test-service", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	testkit.RequireEqual(t, rr.Code, http.StatusOK)

	var s Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	testkit.AssertEqual(t, s.Status, "ok")
	testkit.AssertEqual(t, s.Service, "test-service")
	testkit.AssertEqual(t, s.Version, "v1.0.0")
	testkit.AssertNil(t, s.Checks)
}

func TestHandler_AllHealthy(t *testing.T) {
	db := NewChecker("db", func(ctx context.Context) error { return nil })
	cache := NewChecker("cache", func(ctx context.Context) error { return nil })

	h := Handler("svc", "v2.0.0", db, cache)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	testkit.RequireEqual(t, rr.Code, http.StatusOK)

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	testkit.AssertEqual(t, s.Status, "ok")
	testkit.AssertEqual(t, s.Checks["db"], "ok")
	testkit.AssertEqual(t, s.Checks["cache"], "ok")
}

func TestHandler_Degraded(t *testing.T) {
	healthy := NewChecker("db", func(ctx context.Context) error { return nil })
	unhealthy := NewChecker("cache", func(ctx context.Context) error {
		return errors.New("connection refused")
	})

	h := Handler("svc", "v1.0.0", healthy, unhealthy)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	testkit.RequireEqual(t, rr.Code, http.StatusServiceUnavailable)

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	testkit.AssertEqual(t, s.Status, "degraded")
	testkit.AssertEqual(t, s.Checks["db"], "ok")
	testkit.AssertEqual(t, s.Checks["cache"], "connection refused")
}

func TestHandler_ContentType(t *testing.T) {
	h := Handler("svc", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	ct := rr.Header().Get("Content-Type")
	testkit.AssertEqual(t, ct, "application/json")
}

func TestNewChecker(t *testing.T) {
	c := NewChecker("test", func(ctx context.Context) error { return nil })
	testkit.AssertEqual(t, c.Name(), "test")
	testkit.AssertNoError(t, c.Check(context.Background()))
}

func TestNewChecker_Error(t *testing.T) {
	c := NewChecker("fail", func(ctx context.Context) error {
		return errors.New("down")
	})
	testkit.AssertError(t, c.Check(context.Background()))
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

	testkit.RequireEqual(t, rr.Code, http.StatusOK)
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

	testkit.RequireEqual(t, rr.Code, http.StatusServiceUnavailable)

	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	testkit.AssertEqual(t, s.Checks["db"], "timeout")
	testkit.AssertEqual(t, s.Checks["redis"], "refused")
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
	testkit.AssertEqual(t, s.Status, "degraded")
}

func TestHandler_SingleChecker(t *testing.T) {
	c := NewChecker("single", func(ctx context.Context) error { return nil })
	h := Handler("svc", "v1.0.0", c)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	testkit.RequireEqual(t, rr.Code, http.StatusOK)
	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	testkit.AssertEqual(t, s.Checks["single"], "ok")
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

func TestHandler_PanicChecker(t *testing.T) {
	panicking := NewChecker("panicker", func(ctx context.Context) error {
		panic("something went terribly wrong")
	})
	healthy := NewChecker("db", func(ctx context.Context) error { return nil })

	h := Handler("svc", "v1.0.0", panicking, healthy)
	rr := httptest.NewRecorder()
	// Must not propagate the panic — should return degraded.
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	testkit.RequireEqual(t, rr.Code, http.StatusServiceUnavailable)
	var s Status
	json.NewDecoder(rr.Body).Decode(&s)
	testkit.AssertEqual(t, s.Status, "degraded")
	testkit.AssertEqual(t, s.Checks["db"], "ok")
	testkit.AssertNotEqual(t, s.Checks["panicker"], "")
}

func TestStatus_IsHealthy(t *testing.T) {
	testkit.AssertTrue(t, (Status{Status: "ok"}).IsHealthy())
	testkit.AssertFalse(t, (Status{Status: "degraded"}).IsHealthy())
}

func TestStatus_HasErrors(t *testing.T) {
	testkit.AssertFalse(t, (Status{Status: "ok"}).HasErrors())
	testkit.AssertTrue(t, (Status{Status: "degraded"}).HasErrors())
}

func TestHandler_UptimePresent(t *testing.T) {
	h := Handler("svc", "v1.0.0")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	var s Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	testkit.AssertNotEqual(t, s.Uptime, "")
}
