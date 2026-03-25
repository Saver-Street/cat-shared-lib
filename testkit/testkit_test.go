package testkit

import (
	"errors"
	"fmt"
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

func TestAssertEmpty_Pass(t *testing.T) {
	AssertEmpty(t, "")
}

func TestAssertEmpty_Fail(t *testing.T) {
	mock := &mockT{}
	AssertEmpty(mock, "hello")
	if !mock.errored {
		t.Error("expected failure for non-empty string")
	}
}

func TestAssertNotEmpty_Pass(t *testing.T) {
	AssertNotEmpty(t, "hello")
}

func TestAssertNotEmpty_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNotEmpty(mock, "")
	if !mock.errored {
		t.Error("expected failure for empty string")
	}
}

func TestAssertApprox_Pass(t *testing.T) {
	AssertApprox(t, 1.0, 1.0, 0.001)
	AssertApprox(t, 1.0005, 1.0, 0.001)
}

func TestAssertApprox_Fail(t *testing.T) {
	mock := &mockT{}
	AssertApprox(mock, 1.5, 1.0, 0.001)
	if !mock.errored {
		t.Error("expected failure for values outside epsilon")
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

// ---- AssertNotContains tests ----

func TestAssertNotContains_Pass(t *testing.T) {
	AssertNotContains(t, "hello world", "bye")
}

func TestAssertNotContains_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNotContains(mock, "hello world", "world")
	if !mock.errored {
		t.Error("expected Errorf when substr is present")
	}
}

// ---- AssertMatch tests ----

func TestAssertMatch_Pass(t *testing.T) {
	AssertMatch(t, "abc-123", `^[a-z]+-\d+$`)
	AssertMatch(t, "2024-01-15T10:30:00Z", `\d{4}-\d{2}-\d{2}`)
}

func TestAssertMatch_Fail(t *testing.T) {
	mock := &mockT{}
	AssertMatch(mock, "hello", `^\d+$`)
	if !mock.errored {
		t.Error("expected Errorf for non-matching string")
	}
}

func TestAssertMatch_InvalidRegex(t *testing.T) {
	mock := &mockT{}
	AssertMatch(mock, "hello", `[invalid`)
	if !mock.fatal {
		t.Error("expected Fatalf for invalid regex")
	}
}

// ---- AssertNoMatch tests ----

func TestAssertNoMatch_Pass(t *testing.T) {
	AssertNoMatch(t, "hello", `^\d+$`)
}

func TestAssertNoMatch_Fail(t *testing.T) {
	mock := &mockT{}
	AssertNoMatch(mock, "123", `^\d+$`)
	if !mock.errored {
		t.Error("expected Errorf for matching string")
	}
}

func TestAssertNoMatch_InvalidRegex(t *testing.T) {
	mock := &mockT{}
	AssertNoMatch(mock, "hello", `[invalid`)
	if !mock.fatal {
		t.Error("expected Fatalf for invalid regex")
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
