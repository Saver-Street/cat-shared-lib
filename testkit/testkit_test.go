package testkit

import (
	"bytes"
	"errors"
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

// mockT is a test double for T.
type mockT struct {
	errored bool
	fatal   bool
}

func (m *mockT) Helper()                          {}
func (m *mockT) Errorf(_ string, _ ...any)        { m.errored = true }
func (m *mockT) Fatalf(_ string, _ ...any)        { m.fatal = true; m.errored = true }
