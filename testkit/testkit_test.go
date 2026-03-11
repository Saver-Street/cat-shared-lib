package testkit

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---- T interface tests ----

func TestAssertEqual_Pass(t *testing.T) {
	AssertEqual(t, 42, 42)
	AssertEqual(t, "hello", "hello")
	AssertEqual(t, []int{1, 2}, []int{1, 2})
}

func TestAssertEqual_Fail(t *testing.T) {
	mock := &mockT{}
	AssertEqual(mock, 1, 2)
	if !mock.errored {
		t.Error("expected failure for unequal values")
	}
}

func TestAssertNotEqual_Pass(t *testing.T) {
	AssertNotEqual(t, 1, 2)
}

func TestAssertNotEqual_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNotEqual(mock, 5, 5)
	if !mock.errored {
		t.Error("expected failure for equal values")
	}
}

func TestAssertNil_Pass(t *testing.T) {
	AssertNil(t, nil)
	var p *int
	AssertNil(t, p)
}

func TestAssertNil_Fail(t *testing.T) {
	mock := &mockT{}
	v := 1
	AssertNil(mock, &v)
	if !mock.errored {
		t.Error("expected failure for non-nil")
	}
}

// ---- AssertTrue / AssertFalse tests ----

func TestAssertTrue_Pass(t *testing.T)  { AssertTrue(t, true) }
func TestAssertFalse_Pass(t *testing.T) { AssertFalse(t, false) }

func TestAssertTrue_Fail(t *testing.T) {
	mock := &mockT{}
	AssertTrue(mock, false)
	if !mock.errored {
		t.Error("expected error for false")
	}
}

func TestAssertFalse_Fail(t *testing.T) {
	mock := &mockT{}
	AssertFalse(mock, true)
	if !mock.errored {
		t.Error("expected error for true")
	}
}

// ---- AssertLen tests ----

func TestAssertLen_Slice(t *testing.T)  { AssertLen(t, []int{1, 2, 3}, 3) }
func TestAssertLen_Map(t *testing.T)    { AssertLen(t, map[string]int{"a": 1}, 1) }
func TestAssertLen_String(t *testing.T) { AssertLen(t, "hello", 5) }
func TestAssertLen_Array(t *testing.T)  { AssertLen(t, [2]int{1, 2}, 2) }

func TestAssertLen_Fail(t *testing.T) {
	mock := &mockT{}
	AssertLen(mock, []int{1}, 5)
	if !mock.errored {
		t.Error("expected error for wrong length")
	}
}

func TestAssertLen_UnsupportedType(t *testing.T) {
	mock := &mockT{}
	AssertLen(mock, 42, 1)
	if !mock.fatal {
		t.Error("expected fatal for unsupported type")
	}
}

func TestAssertNoError_Pass(t *testing.T) {
	AssertNoError(t, nil)
}

func TestAssertNoError_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNoError(mock, errors.New("boom"))
	if !mock.fatal {
		t.Error("expected Fatalf for non-nil error")
	}
}

func TestAssertError_Pass(t *testing.T) {
	AssertError(t, errors.New("err"))
}

func TestAssertError_Fail(t *testing.T) {
	mock := &mockT{}
	AssertError(mock, nil)
	if !mock.fatal {
		t.Error("expected Fatalf for nil error")
	}
}

func TestAssertErrorContains_Pass(t *testing.T) {
	AssertErrorContains(t, errors.New("connection refused"), "refused")
}

func TestAssertErrorContains_Nil(t *testing.T) {
	mock := &mockT{}
	AssertErrorContains(mock, nil, "something")
	if !mock.fatal {
		t.Error("expected Fatalf for nil error")
	}
}

func TestAssertErrorContains_Miss(t *testing.T) {
	mock := &mockT{}
	AssertErrorContains(mock, errors.New("timeout"), "refused")
	if !mock.errored {
		t.Error("expected Errorf for missing substr")
	}
}

// ---- AssertErrorIs tests ----

func TestAssertErrorIs_Pass(t *testing.T) {
	wrapped := fmt.Errorf("wrapped: %w", errSentinel)
	AssertErrorIs(t, wrapped, errSentinel)
}

func TestAssertErrorIs_Nil(t *testing.T) {
	mock := &mockT{}
	AssertErrorIs(mock, nil, errSentinel)
	if !mock.fatal {
		t.Error("expected Fatalf for nil error")
	}
}

func TestAssertErrorIs_NoMatch(t *testing.T) {
	mock := &mockT{}
	AssertErrorIs(mock, errors.New("other"), errSentinel)
	if !mock.errored {
		t.Error("expected Errorf for non-matching error")
	}
}

// ---- AssertErrorAs tests ----

func TestAssertErrorAs_Pass(t *testing.T) {
	err := &testCustomError{code: 404}
	wrapped := fmt.Errorf("wrapped: %w", err)
	var target *testCustomError
	AssertErrorAs(t, wrapped, &target)
	if target.code != 404 {
		t.Errorf("expected code 404, got %d", target.code)
	}
}

func TestAssertErrorAs_Nil(t *testing.T) {
	mock := &mockT{}
	var target *testCustomError
	AssertErrorAs(mock, nil, &target)
	if !mock.fatal {
		t.Error("expected Fatalf for nil error")
	}
}

func TestAssertErrorAs_NoMatch(t *testing.T) {
	mock := &mockT{}
	var target *testCustomError
	AssertErrorAs(mock, errors.New("plain"), &target)
	if !mock.errored {
		t.Error("expected Errorf for non-matching error type")
	}
}

func TestAssertErrorAs_BadTarget(t *testing.T) {
	mock := &mockT{}
	AssertErrorAs(mock, errors.New("err"), "not-a-pointer")
	if !mock.fatal {
		t.Error("expected Fatalf for non-pointer target")
	}
}

func TestAssertContains_Pass(t *testing.T) {
	AssertContains(t, "hello world", "world")
}

func TestAssertContains_Fail(t *testing.T) {
	mock := &mockT{}
	AssertContains(mock, "hello", "bye")
	if !mock.errored {
		t.Error("expected Errorf for missing substr")
	}
}

// ---- AssertNotNil tests ----

func TestAssertNotNil_Pass(t *testing.T) {
	v := 42
	AssertNotNil(t, &v)
	AssertNotNil(t, "hello")
}

func TestAssertNotNil_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNotNil(mock, nil)
	if !mock.errored {
		t.Error("expected error for nil value")
	}
}

func TestAssertNotNil_NilPointer(t *testing.T) {
	mock := &mockT{}
	var p *int
	AssertNotNil(mock, p)
	if !mock.errored {
		t.Error("expected error for nil pointer")
	}
}

// ---- AssertPanics tests ----

func TestAssertPanics_Pass(t *testing.T) {
	AssertPanics(t, func() { panic("boom") })
}

func TestAssertPanics_Fail(t *testing.T) {
	mock := &mockT{}
	AssertPanics(mock, func() {})
	if !mock.errored {
		t.Error("expected error when function does not panic")
	}
}

func TestAssertPanicsContains_Pass(t *testing.T) {
	AssertPanicsContains(t, func() { panic("connection refused") }, "refused")
}

func TestAssertPanicsContains_NoPanic(t *testing.T) {
	mock := &mockT{}
	AssertPanicsContains(mock, func() {}, "anything")
	if !mock.errored {
		t.Error("expected error when function does not panic")
	}
}

func TestAssertPanicsContains_WrongMessage(t *testing.T) {
	mock := &mockT{}
	AssertPanicsContains(mock, func() { panic("timeout") }, "refused")
	if !mock.errored {
		t.Error("expected error for wrong panic message")
	}
}

// ---- JSON tests ----

func TestAssertJSON(t *testing.T) {
	type S struct{ X int }
	var got S
	AssertJSON(t, []byte(`{"X":7}`), &got)
	if got.X != 7 {
		t.Errorf("expected X=7, got %d", got.X)
	}
}

func TestAssertJSON_BadJSON(t *testing.T) {
	mock := &mockT{}
	var got map[string]any
	AssertJSON(mock, []byte(`not json`), &got)
	if !mock.fatal {
		t.Error("expected Fatalf for bad JSON")
	}
}

func TestAssertJSONEqual_Pass(t *testing.T) {
	AssertJSONEqual(t, map[string]int{"a": 1}, map[string]int{"a": 1})
}

func TestAssertJSONEqual_Fail(t *testing.T) {
	mock := &mockT{}
	AssertJSONEqual(mock, map[string]int{"a": 1}, map[string]int{"a": 2})
	if !mock.errored {
		t.Error("expected error for JSON mismatch")
	}
}

func TestAssertJSONContains_Pass(t *testing.T) {
	body := []byte(`{"code":"NOT_FOUND","message":"not found","extra":"x"}`)
	AssertJSONContains(t, body, map[string]any{
		"code":    "NOT_FOUND",
		"message": "not found",
	})
}

func TestAssertJSONContains_MissingKey(t *testing.T) {
	mock := &mockT{}
	body := []byte(`{"a":1}`)
	AssertJSONContains(mock, body, map[string]any{"b": 2})
	if !mock.errored {
		t.Error("expected error for missing key")
	}
}

func TestAssertJSONContains_WrongValue(t *testing.T) {
	mock := &mockT{}
	body := []byte(`{"a":1}`)
	AssertJSONContains(mock, body, map[string]any{"a": 2})
	if !mock.errored {
		t.Error("expected error for wrong value")
	}
}

// ---- HTTP helpers tests ----

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

func (m *mockT) Helper()                          {}
func (m *mockT) Errorf(_ string, _ ...any)        { m.errored = true }
func (m *mockT) Fatalf(_ string, _ ...any)        { m.fatal = true; m.errored = true }

// test helpers for AssertErrorIs / AssertErrorAs

var errSentinel = errors.New("sentinel")

type testCustomError struct{ code int }

func (e *testCustomError) Error() string { return fmt.Sprintf("custom error %d", e.code) }
