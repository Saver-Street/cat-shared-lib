package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout_HandlerCompletes(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Errorf("body = %q, want ok:true", rec.Body.String())
	}
}

func TestTimeout_HandlerExceedsDeadline(t *testing.T) {
	handler := Timeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow work by waiting for context cancellation.
		<-r.Context().Done()
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d, want 504", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "timed out") {
		t.Errorf("body = %q, want 'timed out' message", rec.Body.String())
	}
}

func TestTimeout_ZeroDuration_NoOp(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := Timeout(0)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418 (pass-through)", rec.Code)
	}
}

func TestTimeout_NegativeDuration_NoOp(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	handler := Timeout(-1 * time.Second)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202 (pass-through)", rec.Code)
	}
}

func TestTimeout_ContextDeadlineSet(t *testing.T) {
	timeout := 500 * time.Millisecond
	handler := Timeout(timeout)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("expected context to have a deadline")
			return
		}
		remaining := time.Until(deadline)
		if remaining > timeout || remaining <= 0 {
			t.Errorf("deadline remaining = %v, want (0, %v]", remaining, timeout)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestTimeout_HandlerAlreadyWrote(t *testing.T) {
	// Handler starts writing before timeout — timeout should not overwrite.
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("started"))
		// Then sleep past the timeout.
		<-r.Context().Done()
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Handler already wrote 201, so timeout must not overwrite with 504.
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201 (handler already wrote)", rec.Code)
	}
}

func TestTimeoutWriter_WriteSuppressedAfterTimeout(t *testing.T) {
	tw := &timeoutWriter{ResponseWriter: httptest.NewRecorder()}

	// Simulate timeout happening before any writes.
	if !tw.tryTimeout() {
		t.Error("tryTimeout should return true when nothing written")
	}

	// Subsequent writes should be suppressed.
	tw.WriteHeader(http.StatusOK) // should be no-op
	n, err := tw.Write([]byte("data"))
	if n != 0 || err == nil {
		t.Errorf("Write after timeout: n=%d, err=%v; want 0, non-nil", n, err)
	}
}

func TestTimeoutWriter_TryTimeout_AlreadyWrote(t *testing.T) {
	tw := &timeoutWriter{ResponseWriter: httptest.NewRecorder()}
	tw.WriteHeader(http.StatusOK) // handler started

	if tw.tryTimeout() {
		t.Error("tryTimeout should return false when handler already wrote")
	}
}
