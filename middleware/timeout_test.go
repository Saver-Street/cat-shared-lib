package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestTimeout_HandlerCompletes(t *testing.T) {
	handler := Timeout(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	testkit.AssertStatus(t, rec, http.StatusOK)
	testkit.AssertContains(t, rec.Body.String(), `"ok":true`)
}

func TestTimeout_HandlerExceedsDeadline(t *testing.T) {
	handler := Timeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow work by waiting for context cancellation.
		<-r.Context().Done()
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	testkit.AssertStatus(t, rec, http.StatusGatewayTimeout)
	testkit.AssertContains(t, rec.Body.String(), "timed out")
}

func TestTimeout_ZeroDuration_NoOp(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := Timeout(0)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	testkit.AssertStatus(t, rec, http.StatusTeapot)
}

func TestTimeout_NegativeDuration_NoOp(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	handler := Timeout(-1 * time.Second)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	testkit.AssertStatus(t, rec, http.StatusAccepted)
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

	testkit.AssertStatus(t, rec, http.StatusOK)
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
	testkit.AssertStatus(t, rec, http.StatusCreated)
}

func TestTimeoutWriter_WriteSuppressedAfterTimeout(t *testing.T) {
	tw := &timeoutWriter{ResponseWriter: httptest.NewRecorder()}

	// Simulate timeout happening before any writes.
	testkit.AssertTrue(t, tw.tryTimeout())

	// Subsequent writes should be suppressed.
	tw.WriteHeader(http.StatusOK) // should be no-op
	n, err := tw.Write([]byte("data"))
	testkit.AssertEqual(t, n, 0)
	testkit.AssertError(t, err)
}

func TestTimeoutWriter_TryTimeout_AlreadyWrote(t *testing.T) {
	tw := &timeoutWriter{ResponseWriter: httptest.NewRecorder()}
	tw.WriteHeader(http.StatusOK) // handler started

	testkit.AssertFalse(t, tw.tryTimeout())
}
