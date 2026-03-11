// Package testkit provides shared test helpers, assertion utilities, and mock
// implementations for use across service test suites. It is intended to be
// imported only in _test.go files (or test binaries).
package testkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// T is a subset of *testing.T used by helpers so they can accept both
// *testing.T and *testing.B.
type T interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// ---- JSON helpers ----

// AssertJSON unmarshals body into got and calls t.Fatalf on error.
func AssertJSON(t T, body []byte, got any) {
	t.Helper()
	if err := json.Unmarshal(body, got); err != nil {
		t.Fatalf("testkit: unmarshal JSON: %v\nbody: %s", err, body)
	}
}

// AssertJSONEqual asserts that got and want marshal to the same JSON.
func AssertJSONEqual(t T, got, want any) {
	t.Helper()
	g, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("testkit: marshal got: %v", err)
	}
	w, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("testkit: marshal want: %v", err)
	}
	if string(g) != string(w) {
		t.Errorf("testkit: JSON mismatch\n got:  %s\n want: %s", g, w)
	}
}

// AssertJSONContains checks that the JSON object body contains all key-value
// pairs present in subset (also a JSON object).
func AssertJSONContains(t T, body []byte, subset map[string]any) {
	t.Helper()
	var full map[string]any
	if err := json.Unmarshal(body, &full); err != nil {
		t.Fatalf("testkit: unmarshal body: %v", err)
	}
	for k, wantV := range subset {
		gotV, ok := full[k]
		if !ok {
			t.Errorf("testkit: key %q missing from JSON body", k)
			continue
		}
		wantJSON, _ := json.Marshal(wantV)
		gotJSON, _ := json.Marshal(gotV)
		if string(wantJSON) != string(gotJSON) {
			t.Errorf("testkit: key %q: got %s, want %s", k, gotJSON, wantJSON)
		}
	}
}

// ---- Deep equal / struct helpers ----

// AssertEqual fails the test if got != want using reflect.DeepEqual.
func AssertEqual(t T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("testkit: got %v, want %v", got, want)
	}
}

// AssertNotEqual fails the test if got == want.
func AssertNotEqual(t T, got, notWant any) {
	t.Helper()
	if reflect.DeepEqual(got, notWant) {
		t.Errorf("testkit: expected values to differ, both are %v", got)
	}
}

// AssertEmpty fails the test if s is not the empty string.
func AssertEmpty(t T, s string) {
	t.Helper()
	if s != "" {
		t.Errorf("testkit: expected empty string, got %q", s)
	}
}

// AssertNotEmpty fails the test if s is the empty string.
func AssertNotEmpty(t T, s string) {
	t.Helper()
	if s == "" {
		t.Errorf("testkit: expected non-empty string")
	}
}

// AssertApprox fails the test if |got - want| > epsilon.
// Useful for floating-point comparisons that need tolerance.
func AssertApprox(t T, got, want, epsilon float64) {
	t.Helper()
	if math.Abs(got-want) > epsilon {
		t.Errorf("testkit: got %v, want %v (±%v)", got, want, epsilon)
	}
}

// AssertNil fails the test if v is not nil.
func AssertNil(t T, v any) {
	t.Helper()
	if !isNil(v) {
		t.Errorf("testkit: expected nil, got %v", v)
	}
}

// AssertTrue fails the test if v is false.
func AssertTrue(t T, v bool) {
	t.Helper()
	if !v {
		t.Errorf("testkit: expected true, got false")
	}
}

// AssertFalse fails the test if v is true.
func AssertFalse(t T, v bool) {
	t.Helper()
	if v {
		t.Errorf("testkit: expected false, got true")
	}
}

// AssertLen fails the test if the length of v does not equal want.
// v must be a slice, map, string, or channel.
func AssertLen(t T, v any, want int) {
	t.Helper()
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Chan, reflect.Array:
		if rv.Len() != want {
			t.Errorf("testkit: expected length %d, got %d", want, rv.Len())
		}
	default:
		t.Fatalf("testkit: AssertLen: unsupported type %T", v)
	}
}

// AssertNoError fails the test if err != nil.
func AssertNoError(t T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("testkit: unexpected error: %v", err)
	}
}

// AssertError fails the test if err == nil.
func AssertError(t T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("testkit: expected an error, got nil")
	}
}

// AssertErrorContains fails unless err != nil and err.Error() contains substr.
func AssertErrorContains(t T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("testkit: expected error containing %q, got nil", substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("testkit: error %q does not contain %q", err.Error(), substr)
	}
}

// AssertErrorIs fails unless errors.Is(err, target) returns true.
// Use this for sentinel error checks (e.g., io.EOF, context.Canceled).
func AssertErrorIs(t T, err, target error) {
	t.Helper()
	if err == nil {
		t.Fatalf("testkit: expected error matching %v, got nil", target)
		return
	}
	if !errors.Is(err, target) {
		t.Errorf("testkit: error %q is not %v", err.Error(), target)
	}
}

// AssertErrorAs fails unless errors.As(err, target) succeeds.
// target must be a non-nil pointer to the expected error type.
func AssertErrorAs(t T, err error, target any) {
	t.Helper()
	if err == nil {
		t.Fatalf("testkit: expected error assignable to %T, got nil", target)
		return
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		t.Fatalf("testkit: AssertErrorAs target must be a non-nil pointer, got %T", target)
		return
	}
	if !errors.As(err, target) {
		t.Errorf("testkit: error %q is not assignable to %T", err.Error(), target)
	}
}

// AssertContains fails unless s contains substr.
func AssertContains(t T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("testkit: %q does not contain %q", s, substr)
	}
}

// AssertNotContains fails if s contains substr.
func AssertNotContains(t T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("testkit: %q should not contain %q", s, substr)
	}
}

// AssertMatch fails unless s matches the regular expression pattern.
func AssertMatch(t T, s, pattern string) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		t.Fatalf("testkit: invalid regex %q: %v", pattern, err)
		return
	}
	if !matched {
		t.Errorf("testkit: %q does not match pattern %q", s, pattern)
	}
}

// AssertNoMatch fails if s matches the regular expression pattern.
func AssertNoMatch(t T, s, pattern string) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		t.Fatalf("testkit: invalid regex %q: %v", pattern, err)
		return
	}
	if matched {
		t.Errorf("testkit: %q should not match pattern %q", s, pattern)
	}
}

// AssertNotNil fails the test if v is nil.
func AssertNotNil(t T, v any) {
	t.Helper()
	if isNil(v) {
		t.Errorf("testkit: expected non-nil value, got nil")
	}
}

// AssertPanics fails unless fn panics when called.
func AssertPanics(t T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("testkit: expected panic, but function returned normally")
		}
	}()
	fn()
}

// AssertPanicsContains fails unless fn panics with a message containing substr.
func AssertPanicsContains(t T, fn func(), substr string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("testkit: expected panic containing %q, but function returned normally", substr)
			return
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, substr) {
			t.Errorf("testkit: panic %q does not contain %q", msg, substr)
		}
	}()
	fn()
}

// ---- Require helpers (fatal on failure) ----
//
// Require* helpers mirror their Assert* counterparts but call t.Fatalf
// instead of t.Errorf, stopping the test immediately on failure. Use them
// when subsequent test logic depends on the assertion passing (guard checks).

// RequireNoError fails the test immediately if err is non-nil.
func RequireNoError(t T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("testkit: unexpected error: %v", err)
	}
}

// RequireEqual fails the test immediately if got != want (reflect.DeepEqual).
func RequireEqual(t T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("testkit: got %v (%T), want %v (%T)", got, got, want, want)
	}
}

// RequireNil fails the test immediately if v is non-nil.
func RequireNil(t T, v any) {
	t.Helper()
	if !isNil(v) {
		t.Fatalf("testkit: expected nil, got %v", v)
	}
}

// RequireNotNil fails the test immediately if v is nil.
func RequireNotNil(t T, v any) {
	t.Helper()
	if isNil(v) {
		t.Fatalf("testkit: got nil, want non-nil")
	}
}

// RequireLen fails the test immediately if the length of v does not match want.
func RequireLen(t T, v any, want int) {
	t.Helper()
	rv := reflect.ValueOf(v)
	if rv.Len() != want {
		t.Fatalf("testkit: len = %d, want %d", rv.Len(), want)
	}
}

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

// Ptr returns a pointer to v. Useful for creating pointers to literals in
// test table entries (e.g. testkit.Ptr("hello"), testkit.Ptr(int64(42))).
func Ptr[T any](v T) *T { return &v }
