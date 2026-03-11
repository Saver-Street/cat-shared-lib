package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealthResponse(t *testing.T) {
	t.Parallel()
	h := NewHealthResponse(StatusUp)
	if h.Status != StatusUp {
		t.Errorf("Status = %q; want up", h.Status)
	}
	if h.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
	// Verify timestamp is valid RFC3339.
	if _, err := time.Parse(time.RFC3339, h.Timestamp); err != nil {
		t.Errorf("Timestamp = %q; not RFC3339: %v", h.Timestamp, err)
	}
}

func TestHealthResponseWithComponent(t *testing.T) {
	t.Parallel()
	h := NewHealthResponse(StatusUp).
		WithComponent("db", ComponentHealth{Status: StatusUp}).
		WithComponent("cache", ComponentHealth{
			Status:  StatusDegraded,
			Details: map[string]any{"latency_ms": 150},
		})

	if len(h.Components) != 2 {
		t.Fatalf("Components count = %d; want 2", len(h.Components))
	}
	db := h.Components["db"]
	if db.Status != StatusUp {
		t.Errorf("db status = %q; want up", db.Status)
	}
	cache := h.Components["cache"]
	if cache.Status != StatusDegraded {
		t.Errorf("cache status = %q; want degraded", cache.Status)
	}
}

func TestHealthUp(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	h := NewHealthResponse(StatusUp)
	Health(w, h)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", w.Code)
	}
	var got HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != StatusUp {
		t.Errorf("body status = %q; want up", got.Status)
	}
}

func TestHealthDown(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	h := NewHealthResponse(StatusDown)
	Health(w, h)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d; want 503", w.Code)
	}
}

func TestHealthDegraded(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	h := NewHealthResponse(StatusDegraded)
	Health(w, h)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", w.Code)
	}
}

func TestHealthContentType(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	Health(w, NewHealthResponse(StatusUp))
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
}

func TestHealthOmitsEmptyComponents(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	Health(w, NewHealthResponse(StatusUp))

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := raw["components"]; ok {
		t.Error("nil components should be omitted from JSON")
	}
}

func TestHealthConstants(t *testing.T) {
	t.Parallel()
	if StatusUp != "up" {
		t.Errorf("StatusUp = %q", StatusUp)
	}
	if StatusDown != "down" {
		t.Errorf("StatusDown = %q", StatusDown)
	}
	if StatusDegraded != "degraded" {
		t.Errorf("StatusDegraded = %q", StatusDegraded)
	}
}

func BenchmarkHealth(b *testing.B) {
	h := NewHealthResponse(StatusUp).
		WithComponent("db", ComponentHealth{Status: StatusUp})
	for range b.N {
		w := httptest.NewRecorder()
		Health(w, h)
	}
}

func BenchmarkNewHealthResponse(b *testing.B) {
	for range b.N {
		_ = NewHealthResponse(StatusUp)
	}
}

func FuzzHealthStatus(f *testing.F) {
	f.Add("up")
	f.Add("down")
	f.Add("degraded")
	f.Add("")
	f.Fuzz(func(t *testing.T, status string) {
		h := NewHealthResponse(HealthStatus(status))
		w := httptest.NewRecorder()
		Health(w, h)
		if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusOK {
			t.Errorf("unexpected status %d", w.Code)
		}
	})
}
