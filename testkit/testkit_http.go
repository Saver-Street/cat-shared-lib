package testkit

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// ---- HTTP helpers ----

// NewRequest builds an *http.Request for use in handler tests.
// body may be nil.
func NewRequest(method, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(method, target, body)
}

// NewJSONRequest builds an *http.Request with a JSON body and Content-Type header.
func NewJSONRequest(t T, method, target string, body any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("testkit: marshal request body: %v", err)
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// AssertStatus fails unless rr.Code == code.
func AssertStatus(t T, rr *httptest.ResponseRecorder, code int) {
	t.Helper()
	if rr.Code != code {
		t.Errorf("testkit: status %d, want %d\nbody: %s", rr.Code, code, rr.Body)
	}
}

// AssertHeader fails unless rr contains the header key with expected value.
func AssertHeader(t T, rr *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	got := rr.Header().Get(key)
	if got != want {
		t.Errorf("testkit: header %q = %q, want %q", key, got, want)
	}
}

// ---- Mock HTTP server ----

// MockServer is a thin wrapper around httptest.Server with helpers for
// recording requests and setting up response stubs.
type MockServer struct {
	*httptest.Server
	requests []*http.Request
	bodies   [][]byte
	handler  func(w http.ResponseWriter, r *http.Request)
}

// NewMockServer creates a MockServer. Use Handle to set a custom response; by
// default it returns 200 OK with an empty body.
func NewMockServer(t *testing.T) *MockServer {
	ms := &MockServer{}
	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ms.requests = append(ms.requests, r)
		ms.bodies = append(ms.bodies, body)
		if ms.handler != nil {
			// Re-inject body for the custom handler.
			r.Body = io.NopCloser(bytes.NewReader(body))
			ms.handler(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ms.Close)
	return ms
}

// Handle sets a custom handler for all incoming requests.
func (ms *MockServer) Handle(fn func(w http.ResponseWriter, r *http.Request)) {
	ms.handler = fn
}

// RequestCount returns the number of requests received.
func (ms *MockServer) RequestCount() int { return len(ms.requests) }

// LastRequest returns the most recently received request, or nil if none.
func (ms *MockServer) LastRequest() *http.Request {
	if len(ms.requests) == 0 {
		return nil
	}
	return ms.requests[len(ms.requests)-1]
}

// LastBody returns the body of the most recently received request.
func (ms *MockServer) LastBody() []byte {
	if len(ms.bodies) == 0 {
		return nil
	}
	return ms.bodies[len(ms.bodies)-1]
}

// ---- Context helpers ----

// ContextWithValue returns a context carrying the given key/value pair.
func ContextWithValue(parent context.Context, key, val any) context.Context {
	return context.WithValue(parent, key, val)
}

// ---- Mock interfaces ----

// CallRecorder records calls to a function for later assertion.
type CallRecorder struct {
	calls [][]any
}

// Record records a call with the given arguments.
func (cr *CallRecorder) Record(args ...any) {
	cr.calls = append(cr.calls, args)
}

// CallCount returns the number of recorded calls.
func (cr *CallRecorder) CallCount() int { return len(cr.calls) }

// Call returns the arguments of the nth call (0-indexed).
func (cr *CallRecorder) Call(n int) []any {
	if n < 0 || n >= len(cr.calls) {
		return nil
	}
	return cr.calls[n]
}

// AssertCallCount fails unless the recorder has exactly n calls.
func (cr *CallRecorder) AssertCallCount(t T, n int) {
	t.Helper()
	if cr.CallCount() != n {
		t.Errorf("testkit: expected %d call(s), got %d", n, cr.CallCount())
	}
}

// Reset clears all recorded calls.
func (cr *CallRecorder) Reset() { cr.calls = nil }

// ---- Helpers ----

func isNil(v any) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	}
	return false
}

// MustMarshalJSON marshals v to JSON and panics on error. Useful in test setup.
func MustMarshalJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("testkit: MustMarshalJSON: %v", err))
	}
	return b
}

// AssertGreater asserts that got > want.
func AssertGreater[V cmp.Ordered](t T, got, want V) {
	t.Helper()
	if got <= want {
		t.Errorf("expected %v > %v", got, want)
	}
}

// AssertLess asserts that got < want.
func AssertLess[V cmp.Ordered](t T, got, want V) {
	t.Helper()
	if got >= want {
		t.Errorf("expected %v < %v", got, want)
	}
}

// AssertHasPrefix asserts that s starts with the given prefix.
func AssertHasPrefix(t T, s, prefix string) {
	t.Helper()
	if !strings.HasPrefix(s, prefix) {
		t.Errorf("expected %q to have prefix %q", s, prefix)
	}
}

// AssertHasSuffix asserts that s ends with the given suffix.
func AssertHasSuffix(t T, s, suffix string) {
	t.Helper()
	if !strings.HasSuffix(s, suffix) {
		t.Errorf("expected %q to have suffix %q", s, suffix)
	}
}

// AssertMapHasKey asserts that the map contains the given key.
func AssertMapHasKey[K comparable, V any](t T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; !ok {
		t.Errorf("expected map to contain key %v", key)
	}
}

// AssertMapNotHasKey asserts that the map does not contain the given key.
func AssertMapNotHasKey[K comparable, V any](t T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("expected map not to contain key %v", key)
	}
}

// Ptr returns a pointer to v. Useful for creating pointers to literals in
// test table entries (e.g. testkit.Ptr("hello"), testkit.Ptr(int64(42))).
func Ptr[T any](v T) *T { return &v }

// AssertWithin asserts that a duration is at most maxDur. Useful for
// verifying timeouts, latency bounds, and retry back-off windows.
func AssertWithin(t T, got, maxDur time.Duration) {
	t.Helper()
	if got > maxDur {
		t.Errorf("expected duration ≤ %v, got %v", maxDur, got)
	}
}

// AssertBetween asserts that got is in the inclusive range [lo, hi].
func AssertBetween[V cmp.Ordered](t T, got, lo, hi V) {
	t.Helper()
	if got < lo || got > hi {
		t.Errorf("expected %v to be between %v and %v", got, lo, hi)
	}
}

// ---------------------------------------------------------------------------
// Fixture & file helpers
// ---------------------------------------------------------------------------

// LoadFixture reads a file relative to the test's working directory and returns
// its contents as bytes. It calls t.Fatal on any read error.
func LoadFixture(t T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("LoadFixture: %v", err)
	}
	return data
}

// LoadJSONFixture reads a JSON file and unmarshals it into dst.
func LoadJSONFixture(t T, path string, dst any) {
	t.Helper()
	data := LoadFixture(t, path)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("LoadJSONFixture: unmarshal %s: %v", path, err)
	}
}

// WriteFixture creates a temporary file with the given content in dir and
// returns its path. The file is automatically removed when the test finishes.
func WriteFixture(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatalf("WriteFixture: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(p) })
	return p
}

// TempDir creates a temporary directory that is automatically removed when
// the test finishes and returns its path.
func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// TempFile creates a named file with the given content inside a temporary
// directory. The directory is automatically removed when the test finishes.
func TempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	return p
}

// AssertFileExists fails if the file at path does not exist.
func AssertFileExists(t T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileContains fails if the file content does not contain substr.
func AssertFileContains(t T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("AssertFileContains: %v", err)
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("expected file %s to contain %q", path, substr)
	}
}

// ---------------------------------------------------------------------------
// Slice assertions
// ---------------------------------------------------------------------------

// AssertSliceContains fails if the slice does not contain the element.
func AssertSliceContains[V comparable](t T, s []V, elem V) {
	t.Helper()
	if !slices.Contains(s, elem) {
		t.Errorf("expected slice to contain %v, got %v", elem, s)
	}
}

// AssertSliceNotContains fails if the slice contains the element.
func AssertSliceNotContains[V comparable](t T, s []V, elem V) {
	t.Helper()
	if slices.Contains(s, elem) {
		t.Errorf("expected slice not to contain %v", elem)
	}
}

// AssertSliceEqual fails if two slices are not deeply equal.
func AssertSliceEqual[V comparable](t T, got, want []V) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("slice mismatch\ngot:  %v\nwant: %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Async / eventual consistency helper
// ---------------------------------------------------------------------------

// Eventually retries fn until it returns nil or the timeout is reached.
// The check interval defaults to 10ms. Use this for assertions on
// asynchronous or eventually-consistent state.
func Eventually(t T, timeout time.Duration, fn func() error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	interval := 10 * time.Millisecond
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = fn()
		if lastErr == nil {
			return
		}
		time.Sleep(interval)
	}
	t.Errorf("Eventually timed out after %v: %v", timeout, lastErr)
}
