package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHandler_ConcurrentRequests(t *testing.T) {
	const checkerCount = 20
	var checks []Checker
	for i := range checkerCount {
		name := "checker-" + string(rune('A'+i))
		checks = append(checks, NewChecker(name, func(_ context.Context) error {
			return nil
		}))
	}
	handler := Handler("svc", "1.0", checks...)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			r := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			var status Status
			if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
				t.Errorf("decode error: %v", err)
				return
			}
			if status.Status != "ok" {
				t.Errorf("status = %q, want ok", status.Status)
			}
		}()
	}
	wg.Wait()
}

func TestHandler_DegradedChecker(t *testing.T) {
	checker := NewChecker("failing", func(_ context.Context) error {
		return errors.New("disk full")
	})
	handler := Handler("svc", "1.0", checker)

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != 503 {
		t.Errorf("status code = %d, want 503", w.Code)
	}
	var status Status
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if status.Status != "degraded" {
		t.Errorf("status = %q, want degraded", status.Status)
	}
}

func FuzzHandler(f *testing.F) {
	f.Add("myservice", "1.0.0")
	f.Add("", "")
	f.Add("a", "b")
	f.Add("service-with-dashes", "v2.3.4-beta")

	f.Fuzz(func(t *testing.T, service, version string) {
		handler := Handler(service, version)
		r := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		var status Status
		if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if status.Service != service {
			t.Errorf("service = %q, want %q", status.Service, service)
		}
		if status.Version != version {
			t.Errorf("version = %q, want %q", status.Version, version)
		}
		if status.Status != "ok" {
			t.Errorf("status = %q, want ok (no checkers)", status.Status)
		}
	})
}
