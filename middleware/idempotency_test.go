package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIdempotency_POST_CachesResponse(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))

	// First request
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "abc-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first call status = %d, want 201", rec.Code)
	}
	if rec.Body.String() != "created" {
		t.Fatalf("first call body = %q", rec.Body.String())
	}

	// Second request with same key
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Idempotency-Key", "abc-123")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Fatalf("second call status = %d, want 201", rec2.Code)
	}
	if rec2.Body.String() != "created" {
		t.Fatalf("second call body = %q", rec2.Body.String())
	}
	if rec2.Header().Get("X-Custom") != "value" {
		t.Fatalf("second call header X-Custom = %q", rec2.Header().Get("X-Custom"))
	}
	if callCount != 1 {
		t.Fatalf("handler called %d times, want 1", callCount)
	}
}

func TestIdempotency_GET_PassesThrough(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("ok"))
	}))

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Idempotency-Key", "same-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 3 {
		t.Fatalf("GET handler called %d times, want 3", callCount)
	}
}

func TestIdempotency_DELETE_PassesThrough(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
	}))

	for range 2 {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req.Header.Set("Idempotency-Key", "del-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 2 {
		t.Fatalf("DELETE handler called %d times, want 2", callCount)
	}
}

func TestIdempotency_NoHeader(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
	}))

	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 2 {
		t.Fatalf("handler called %d times without idempotency header, want 2", callCount)
	}
}

func TestIdempotency_CustomHeader(t *testing.T) {
	handler := Idempotency(
		WithIdempotencyHeader("X-Request-Id"),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Request-Id", "req-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Replay
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("X-Request-Id", "req-1")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Body.String() != "ok" {
		t.Fatalf("replay body = %q, want %q", rec2.Body.String(), "ok")
	}
}

func TestIdempotency_PUT(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("updated"))
	}))

	for range 2 {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		req.Header.Set("Idempotency-Key", "put-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 1 {
		t.Fatalf("PUT handler called %d times, want 1", callCount)
	}
}

func TestIdempotency_PATCH(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte("patched"))
	}))

	for range 2 {
		req := httptest.NewRequest(http.MethodPatch, "/", nil)
		req.Header.Set("Idempotency-Key", "patch-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 1 {
		t.Fatalf("PATCH handler called %d times, want 1", callCount)
	}
}

func TestIdempotency_DifferentKeys(t *testing.T) {
	callCount := 0
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
	}))

	for _, key := range []string{"a", "b", "c"} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", key)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	if callCount != 3 {
		t.Fatalf("handler called %d times, want 3 (different keys)", callCount)
	}
}

func TestIdempotency_CustomStore(t *testing.T) {
	store := NewMemoryIdempotencyStore(time.Hour)
	handler := Idempotency(WithIdempotencyStore(store))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("stored"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "custom-store")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	status, _, body, ok := store.Get("custom-store")
	if !ok {
		t.Fatal("expected entry in custom store")
	}
	if status != http.StatusOK {
		t.Fatalf("stored status = %d, want 200", status)
	}
	if string(body) != "stored" {
		t.Fatalf("stored body = %q", string(body))
	}
}

func TestMemoryIdempotencyStore_TTLExpired(t *testing.T) {
	store := NewMemoryIdempotencyStore(1 * time.Millisecond)
	store.Set("key", 200, nil, []byte("data"))
	time.Sleep(5 * time.Millisecond)

	_, _, _, ok := store.Get("key")
	if ok {
		t.Fatal("expected entry to be expired")
	}
}

func TestMemoryIdempotencyStore_NotFound(t *testing.T) {
	store := NewMemoryIdempotencyStore(time.Hour)
	_, _, _, ok := store.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRequiresIdempotency(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodGet, false},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, false},
		{http.MethodHead, false},
		{http.MethodOptions, false},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if got := requiresIdempotency(tt.method); got != tt.want {
				t.Fatalf("requiresIdempotency(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestIdempotency_DefaultStatusCode(t *testing.T) {
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't call WriteHeader; rely on default 200.
		w.Write([]byte("default"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "default-status")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Replay
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Idempotency-Key", "default-status")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200", rec2.Code)
	}
}

func BenchmarkIdempotency_NewRequest(b *testing.B) {
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "bench-new")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkIdempotency_CachedReplay(b *testing.B) {
	handler := Idempotency()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// Seed the cache.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Idempotency-Key", "bench-cached")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Idempotency-Key", "bench-cached")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

// Suppress unused import.
var _ = io.Discard
