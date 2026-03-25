package testkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequireNoError_Pass(t *testing.T) {
	RequireNoError(t, nil)
}

func TestRequireNoError_Fail(t *testing.T) {
	mt := &mockT{}
	RequireNoError(mt, errors.New("fail"))
	if !mt.fatal {
		t.Error("expected Fatalf")
	}
}

func TestRequireEqual_Pass(t *testing.T) {
	RequireEqual(t, 42, 42)
	RequireEqual(t, "hello", "hello")
}

func TestRequireEqual_Fail(t *testing.T) {
	mt := &mockT{}
	RequireEqual(mt, 1, 2)
	if !mt.fatal {
		t.Error("expected Fatalf")
	}
}

func TestRequireNil_Pass(t *testing.T) {
	RequireNil(t, nil)
}

func TestRequireNil_Fail(t *testing.T) {
	mt := &mockT{}
	RequireNil(mt, "non-nil")
	if !mt.fatal {
		t.Error("expected Fatalf")
	}
}

func TestRequireNotNil_Pass(t *testing.T) {
	RequireNotNil(t, "something")
}

func TestRequireNotNil_Fail(t *testing.T) {
	mt := &mockT{}
	RequireNotNil(mt, nil)
	if !mt.fatal {
		t.Error("expected Fatalf")
	}
}

func TestRequireLen_Pass(t *testing.T) {
	RequireLen(t, []int{1, 2, 3}, 3)
}

func TestRequireLen_Fail(t *testing.T) {
	mt := &mockT{}
	RequireLen(mt, []int{1}, 5)
	if !mt.fatal {
		t.Error("expected Fatalf")
	}
}

func TestPtr_String(t *testing.T) {
	p := Ptr("hello")
	AssertEqual(t, *p, "hello")
}

func TestPtr_Int(t *testing.T) {
	p := Ptr(42)
	AssertEqual(t, *p, 42)
}

func TestPtr_Bool(t *testing.T) {
	p := Ptr(true)
	AssertTrue(t, *p)
}

func TestAssertGreater_Pass(t *testing.T) {
	mt := &mockT{}
	AssertGreater(mt, 10, 5)
	AssertFalse(t, mt.errored)
}

func TestAssertGreater_Fail_Equal(t *testing.T) {
	mt := &mockT{}
	AssertGreater(mt, 5, 5)
	AssertTrue(t, mt.errored)
}

func TestAssertGreater_Fail_Less(t *testing.T) {
	mt := &mockT{}
	AssertGreater(mt, 3, 5)
	AssertTrue(t, mt.errored)
}

func TestAssertLess_Pass(t *testing.T) {
	mt := &mockT{}
	AssertLess(mt, 3, 5)
	AssertFalse(t, mt.errored)
}

func TestAssertLess_Fail_Equal(t *testing.T) {
	mt := &mockT{}
	AssertLess(mt, 5, 5)
	AssertTrue(t, mt.errored)
}

func TestAssertLess_Fail_Greater(t *testing.T) {
	mt := &mockT{}
	AssertLess(mt, 10, 5)
	AssertTrue(t, mt.errored)
}

func TestAssertGreater_String(t *testing.T) {
	mt := &mockT{}
	AssertGreater(mt, "b", "a")
	AssertFalse(t, mt.errored)
}

func TestAssertLess_String(t *testing.T) {
	mt := &mockT{}
	AssertLess(mt, "a", "b")
	AssertFalse(t, mt.errored)
}

func TestAssertHasPrefix_Pass(t *testing.T) {
	mt := &mockT{}
	AssertHasPrefix(mt, "hello world", "hello")
	AssertFalse(t, mt.errored)
}

func TestAssertHasPrefix_Fail(t *testing.T) {
	mt := &mockT{}
	AssertHasPrefix(mt, "hello world", "world")
	AssertTrue(t, mt.errored)
}

func TestAssertHasSuffix_Pass(t *testing.T) {
	mt := &mockT{}
	AssertHasSuffix(mt, "hello world", "world")
	AssertFalse(t, mt.errored)
}

func TestAssertHasSuffix_Fail(t *testing.T) {
	mt := &mockT{}
	AssertHasSuffix(mt, "hello world", "hello")
	AssertTrue(t, mt.errored)
}

func TestAssertMapHasKey_Pass(t *testing.T) {
	mt := &mockT{}
	m := map[string]int{"a": 1, "b": 2}
	AssertMapHasKey(mt, m, "a")
	AssertFalse(t, mt.errored)
}

func TestAssertMapHasKey_Fail(t *testing.T) {
	mt := &mockT{}
	m := map[string]int{"a": 1}
	AssertMapHasKey(mt, m, "z")
	AssertTrue(t, mt.errored)
}

func TestAssertMapNotHasKey_Pass(t *testing.T) {
	mt := &mockT{}
	m := map[string]int{"a": 1}
	AssertMapNotHasKey(mt, m, "z")
	AssertFalse(t, mt.errored)
}

func TestAssertMapNotHasKey_Fail(t *testing.T) {
	mt := &mockT{}
	m := map[string]int{"a": 1}
	AssertMapNotHasKey(mt, m, "a")
	AssertTrue(t, mt.errored)
}

func TestAssertMapHasKey_IntKey(t *testing.T) {
	mt := &mockT{}
	m := map[int]string{42: "answer"}
	AssertMapHasKey(mt, m, 42)
	AssertFalse(t, mt.errored)
}

func TestAssertWithin_Pass(t *testing.T) {
	m := &mockT{}
	AssertWithin(m, 50*time.Millisecond, 100*time.Millisecond)
	if m.errored {
		t.Fatal("expected no error for duration within bounds")
	}
}

func TestAssertWithin_Fail(t *testing.T) {
	m := &mockT{}
	AssertWithin(m, 200*time.Millisecond, 100*time.Millisecond)
	if !m.errored {
		t.Fatal("expected error for duration exceeding bounds")
	}
}

func TestAssertWithin_Equal(t *testing.T) {
	m := &mockT{}
	AssertWithin(m, 100*time.Millisecond, 100*time.Millisecond)
	if m.errored {
		t.Fatal("expected no error for duration equal to bound")
	}
}

func TestAssertBetween_Pass(t *testing.T) {
	m := &mockT{}
	AssertBetween(m, 5, 1, 10)
	if m.errored {
		t.Fatal("expected no error for value in range")
	}
}

func TestAssertBetween_Fail_Low(t *testing.T) {
	m := &mockT{}
	AssertBetween(m, 0, 1, 10)
	if !m.errored {
		t.Fatal("expected error for value below range")
	}
}

func TestAssertBetween_Fail_High(t *testing.T) {
	m := &mockT{}
	AssertBetween(m, 11, 1, 10)
	if !m.errored {
		t.Fatal("expected error for value above range")
	}
}

func TestAssertBetween_Boundary(t *testing.T) {
	m := &mockT{}
	AssertBetween(m, 1, 1, 10)
	AssertBetween(m, 10, 1, 10)
	if m.errored {
		t.Fatal("expected no error for boundary values")
	}
}

func TestRequireTrue_Pass(t *testing.T) {
	m := &mockT{}
	RequireTrue(m, true)
	if m.fatal {
		t.Fatal("expected no fatal for true")
	}
}

func TestRequireTrue_Fail(t *testing.T) {
	m := &mockT{}
	RequireTrue(m, false)
	if !m.fatal {
		t.Fatal("expected fatal for false")
	}
}

func TestRequireFalse_Pass(t *testing.T) {
	m := &mockT{}
	RequireFalse(m, false)
	if m.fatal {
		t.Fatal("expected no fatal for false")
	}
}

func TestRequireFalse_Fail(t *testing.T) {
	m := &mockT{}
	RequireFalse(m, true)
	if !m.fatal {
		t.Fatal("expected fatal for true")
	}
}

// ---------------------------------------------------------------------------
// Fixture & file helpers
// ---------------------------------------------------------------------------

func TestLoadFixture(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "data.txt")
	os.WriteFile(p, []byte("hello"), 0o644)

	data := LoadFixture(t, p)
	AssertEqual(t, string(data), "hello")
}

func TestLoadJSONFixture(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "data.json")
	os.WriteFile(p, []byte(`{"name":"cat"}`), 0o644)

	var got map[string]string
	LoadJSONFixture(t, p, &got)
	AssertEqual(t, got["name"], "cat")
}

func TestWriteFixture(t *testing.T) {
	dir := t.TempDir()
	p := WriteFixture(t, dir, "test.txt", []byte("content"))
	data, err := os.ReadFile(p)
	AssertNoError(t, err)
	AssertEqual(t, string(data), "content")
}

func TestTempFile(t *testing.T) {
	p := TempFile(t, "f.txt", []byte("tmp"))
	data, err := os.ReadFile(p)
	AssertNoError(t, err)
	AssertEqual(t, string(data), "tmp")
}

func TestAssertFileExists(t *testing.T) {
	p := TempFile(t, "exists.txt", []byte("ok"))
	AssertFileExists(t, p)
}

func TestAssertFileExists_Fail(t *testing.T) {
	m := &mockT{}
	AssertFileExists(m, "/no/such/file/xyz")
	if !m.errored {
		t.Fatal("expected failure for missing file")
	}
}

func TestAssertFileContains(t *testing.T) {
	p := TempFile(t, "fc.txt", []byte("hello world"))
	AssertFileContains(t, p, "world")
}

func TestAssertFileContains_Fail(t *testing.T) {
	p := TempFile(t, "fc2.txt", []byte("hello"))
	m := &mockT{}
	AssertFileContains(m, p, "missing")
	if !m.errored {
		t.Fatal("expected failure for missing substring")
	}
}

// ---------------------------------------------------------------------------
// Slice assertions
// ---------------------------------------------------------------------------

func TestAssertSliceContains(t *testing.T) {
	AssertSliceContains(t, []int{1, 2, 3}, 2)
}

func TestAssertSliceContains_Fail(t *testing.T) {
	m := &mockT{}
	AssertSliceContains(m, []int{1, 2, 3}, 99)
	if !m.errored {
		t.Fatal("expected failure")
	}
}

func TestAssertSliceNotContains(t *testing.T) {
	AssertSliceNotContains(t, []int{1, 2, 3}, 99)
}

func TestAssertSliceNotContains_Fail(t *testing.T) {
	m := &mockT{}
	AssertSliceNotContains(m, []int{1, 2, 3}, 2)
	if !m.errored {
		t.Fatal("expected failure")
	}
}

func TestAssertSliceEqual(t *testing.T) {
	AssertSliceEqual(t, []string{"a", "b"}, []string{"a", "b"})
}

func TestAssertSliceEqual_Fail(t *testing.T) {
	m := &mockT{}
	AssertSliceEqual(m, []string{"a"}, []string{"b"})
	if !m.errored {
		t.Fatal("expected failure")
	}
}

// ---------------------------------------------------------------------------
// Eventually
// ---------------------------------------------------------------------------

func TestEventually_Pass(t *testing.T) {
	var counter atomic.Int32
	go func() {
		for range 5 {
			counter.Add(1)
			<-time.After(5 * time.Millisecond)
		}
	}()
	Eventually(t, time.Second, func() error {
		if counter.Load() < 3 {
			return fmt.Errorf("counter is %d, want >= 3", counter.Load())
		}
		return nil
	})
}

func TestEventually_Timeout(t *testing.T) {
	m := &mockT{}
	Eventually(m, 50*time.Millisecond, func() error {
		return fmt.Errorf("always fails")
	})
	if !m.errored {
		t.Fatal("expected failure on timeout")
	}
}

// ---------------------------------------------------------------------------
// LoadJSONFixture with struct
// ---------------------------------------------------------------------------

func TestLoadJSONFixture_Struct(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	dir := t.TempDir()
	data, _ := json.Marshal(item{ID: 1, Name: "cat"})
	p := filepath.Join(dir, "item.json")
	os.WriteFile(p, data, 0o644)

	var got item
	LoadJSONFixture(t, p, &got)
	AssertEqual(t, got.ID, 1)
	AssertEqual(t, got.Name, "cat")
}
