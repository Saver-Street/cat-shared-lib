package testkit

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("GET", "/foo", nil)
	if req.Method != "GET" || req.URL.Path != "/foo" {
		t.Errorf("unexpected request: %+v", req)
	}
}

func TestNewJSONRequest(t *testing.T) {
	req := NewJSONRequest(t, "POST", "/bar", map[string]string{"key": "val"})
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("expected JSON content-type")
	}
}

func TestAssertStatus_Pass(t *testing.T) {
	rr := httptest.NewRecorder()
	rr.WriteHeader(http.StatusCreated)
	AssertStatus(t, rr, http.StatusCreated)
}

func TestAssertStatus_Fail(t *testing.T) {
	mock := &mockT{}
	rr := httptest.NewRecorder()
	rr.WriteHeader(http.StatusNotFound)
	AssertStatus(mock, rr, http.StatusOK)
	if !mock.errored {
		t.Error("expected error for status mismatch")
	}
}

func TestAssertHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rr.Header().Set("X-Foo", "bar")
	AssertHeader(t, rr, "X-Foo", "bar")
}

func TestAssertHeader_Fail(t *testing.T) {
	mock := &mockT{}
	rr := httptest.NewRecorder()
	AssertHeader(mock, rr, "X-Missing", "something")
	if !mock.errored {
		t.Error("expected error for missing header")
	}
}

// ---- MockServer tests ----

func TestMockServer_Basic(t *testing.T) {
	ms := NewMockServer(t)
	resp, err := http.Get(ms.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if ms.RequestCount() != 1 {
		t.Errorf("expected 1 request, got %d", ms.RequestCount())
	}
}

func TestMockServer_CustomHandler(t *testing.T) {
	ms := NewMockServer(t)
	ms.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	resp, err := http.Get(ms.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("expected 418, got %d", resp.StatusCode)
	}
}

func TestMockServer_LastRequest(t *testing.T) {
	ms := NewMockServer(t)
	if ms.LastRequest() != nil {
		t.Error("expected nil before any requests")
	}
	resp, err := http.Get(ms.URL + "/ping") //nolint:noctx
	if err == nil {
		_ = resp.Body.Close()
	}
	if ms.LastRequest() == nil {
		t.Fatal("expected non-nil request")
	}
}

func TestMockServer_LastBody(t *testing.T) {
	ms := NewMockServer(t)
	if ms.LastBody() != nil {
		t.Error("expected nil before any requests")
	}
	resp, err := http.Post(ms.URL, "application/json", bytes.NewBufferString(`{"x":1}`)) //nolint:noctx
	if err == nil {
		_ = resp.Body.Close()
	}
	if string(ms.LastBody()) != `{"x":1}` {
		t.Errorf("unexpected body: %s", ms.LastBody())
	}
}

// ---- CallRecorder tests ----

func TestCallRecorder_Basic(t *testing.T) {
	cr := &CallRecorder{}
	cr.Record("a", 1)
	cr.Record("b", 2)
	cr.AssertCallCount(t, 2)
	if cr.Call(0)[0] != "a" {
		t.Error("expected first call arg 'a'")
	}
}

func TestCallRecorder_Reset(t *testing.T) {
	cr := &CallRecorder{}
	cr.Record("x")
	cr.Reset()
	cr.AssertCallCount(t, 0)
}

func TestCallRecorder_CallOutOfRange(t *testing.T) {
	cr := &CallRecorder{}
	if cr.Call(0) != nil {
		t.Error("expected nil for out-of-range call")
	}
}

func TestCallRecorder_AssertFail(t *testing.T) {
	mock := &mockT{}
	cr := &CallRecorder{}
	cr.AssertCallCount(mock, 5)
	if !mock.errored {
		t.Error("expected error for wrong call count")
	}
}

// ---- Misc ----

func TestMustMarshalJSON(t *testing.T) {
	b := MustMarshalJSON(map[string]int{"x": 1})
	if string(b) != `{"x":1}` {
		t.Errorf("unexpected: %s", b)
	}
}

func TestMustMarshalJSON_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for un-marshalable value")
		}
	}()
	MustMarshalJSON(make(chan int))
}

func TestIsNil_Variants(t *testing.T) {
	if !isNil(nil) {
		t.Error("nil should be nil")
	}
	var p *int
	if !isNil(p) {
		t.Error("nil pointer should be nil")
	}
	v := 1
	if isNil(&v) {
		t.Error("non-nil pointer should not be nil")
	}
}

// ---- ContextWithValue tests ----

func TestContextWithValue(t *testing.T) {
	type ctxKey string
	ctx := ContextWithValue(t.Context(), ctxKey("role"), "admin")
	if got := ctx.Value(ctxKey("role")); got != "admin" {
		t.Errorf("ContextWithValue: got %v, want admin", got)
	}
}

// ---- NewJSONRequest error path ----

func TestNewJSONRequest_MarshalError(t *testing.T) {
	mock := &mockT{}
	_ = NewJSONRequest(mock, "POST", "/", make(chan int))
	if !mock.fatal {
		t.Error("expected Fatalf for un-marshalable body")
	}
}

// ---- AssertJSONEqual error paths ----

func TestAssertJSONEqual_MarshalGotError(t *testing.T) {
	mock := &mockT{}
	AssertJSONEqual(mock, make(chan int), map[string]int{"a": 1})
	if !mock.fatal {
		t.Error("expected Fatalf when got is un-marshalable")
	}
}

func TestAssertJSONEqual_MarshalWantError(t *testing.T) {
	mock := &mockT{}
	AssertJSONEqual(mock, map[string]int{"a": 1}, make(chan int))
	if !mock.fatal {
		t.Error("expected Fatalf when want is un-marshalable")
	}
}

// ---- AssertJSONContains error paths ----

func TestAssertJSONContains_InvalidJSON(t *testing.T) {
	mock := &mockT{}
	AssertJSONContains(mock, []byte("not json"), map[string]any{"x": 1})
	if !mock.fatal {
		t.Error("expected Fatalf for invalid JSON body")
	}
}

// ---- isNil additional type coverage ----

func TestIsNil_NilSlice(t *testing.T) {
	var s []int
	if !isNil(s) {
		t.Error("nil slice should be nil")
	}
	s = []int{1}
	if isNil(s) {
		t.Error("non-nil slice should not be nil")
	}
}

func TestIsNil_NilMap(t *testing.T) {
	var m map[string]int
	if !isNil(m) {
		t.Error("nil map should be nil")
	}
	m = map[string]int{"a": 1}
	if isNil(m) {
		t.Error("non-nil map should not be nil")
	}
}

func TestIsNil_NilFunc(t *testing.T) {
	var fn func()
	if !isNil(fn) {
		t.Error("nil func should be nil")
	}
	fn = func() {}
	if isNil(fn) {
		t.Error("non-nil func should not be nil")
	}
}

func TestIsNil_NilChan(t *testing.T) {
	var ch chan int
	if !isNil(ch) {
		t.Error("nil chan should be nil")
	}
	ch = make(chan int)
	if isNil(ch) {
		t.Error("non-nil chan should not be nil")
	}
}

func TestIsNil_NonNilableTypes(t *testing.T) {
	// int, string, struct should never be nil.
	if isNil(0) {
		t.Error("int should not be nil")
	}
	if isNil("") {
		t.Error("empty string should not be nil")
	}
	if isNil(struct{}{}) {
		t.Error("struct should not be nil")
	}
}

// mockT is a test double for T.
type mockT struct {
	errored bool
	fatal   bool
}

func (m *mockT) Helper()                   {}
func (m *mockT) Errorf(_ string, _ ...any) { m.errored = true }
func (m *mockT) Fatalf(_ string, _ ...any) { m.fatal = true; m.errored = true }

// test helpers for AssertErrorIs / AssertErrorAs

var errSentinel = errors.New("sentinel")

type testCustomError struct{ code int }

func (e *testCustomError) Error() string { return fmt.Sprintf("custom error %d", e.code) }

// ---- Require helpers tests ----
