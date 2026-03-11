package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHandler_ConcurrentRequests(t *testing.T) {
	const checkerCount = 20
	checks := make([]Checker, 0, checkerCount)
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
			err := json.NewDecoder(w.Body).Decode(&status)
			testkit.AssertNoError(t, err)
			testkit.AssertEqual(t, status.Status, "ok")
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

	testkit.AssertEqual(t, w.Code, 503)
	var status Status
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	testkit.AssertEqual(t, status.Status, "degraded")
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
		testkit.AssertEqual(t, status.Service, service)
		testkit.AssertEqual(t, status.Version, version)
		testkit.AssertEqual(t, status.Status, "ok")
	})
}
