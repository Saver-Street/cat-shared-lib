package types

import "testing"

func TestTuple3(t *testing.T) {
	t.Parallel()
	tup := NewTuple3("hello", 42, true)
	if tup.First != "hello" || tup.Second != 42 || tup.Third != true {
		t.Errorf("Tuple3 = %v", tup)
	}
}

func TestTuple3Unpack(t *testing.T) {
	t.Parallel()
	a, b, c := NewTuple3("x", 1, 2.0).Unpack()
	if a != "x" || b != 1 || c != 2.0 {
		t.Errorf("Unpack = %v, %v, %v", a, b, c)
	}
}

func TestTuple3ZeroValue(t *testing.T) {
	t.Parallel()
	var tup Tuple3[string, int, bool]
	if tup.First != "" || tup.Second != 0 || tup.Third != false {
		t.Errorf("zero Tuple3 = %v", tup)
	}
}

func TestTuple4(t *testing.T) {
	t.Parallel()
	tup := NewTuple4("a", 1, true, 3.14)
	if tup.First != "a" || tup.Second != 1 || tup.Third != true || tup.Fourth != 3.14 {
		t.Errorf("Tuple4 = %v", tup)
	}
}

func TestTuple4Unpack(t *testing.T) {
	t.Parallel()
	a, b, c, d := NewTuple4("x", 1, 2.0, false).Unpack()
	if a != "x" || b != 1 || c != 2.0 || d != false {
		t.Errorf("Unpack = %v, %v, %v, %v", a, b, c, d)
	}
}

func TestTuple4ZeroValue(t *testing.T) {
	t.Parallel()
	var tup Tuple4[string, int, bool, float64]
	if tup.First != "" || tup.Second != 0 || tup.Third != false || tup.Fourth != 0.0 {
		t.Errorf("zero Tuple4 = %v", tup)
	}
}

func TestTuple5(t *testing.T) {
	t.Parallel()
	tup := NewTuple5("a", 1, true, 3.14, byte(0xFF))
	if tup.First != "a" || tup.Second != 1 || tup.Third != true || tup.Fourth != 3.14 || tup.Fifth != byte(0xFF) {
		t.Errorf("Tuple5 = %v", tup)
	}
}

func TestTuple5Unpack(t *testing.T) {
	t.Parallel()
	a, b, c, d, e := NewTuple5("x", 1, 2.0, false, "y").Unpack()
	if a != "x" || b != 1 || c != 2.0 || d != false || e != "y" {
		t.Errorf("Unpack = %v, %v, %v, %v, %v", a, b, c, d, e)
	}
}

func TestTuple5ZeroValue(t *testing.T) {
	t.Parallel()
	var tup Tuple5[string, int, bool, float64, byte]
	if tup.First != "" || tup.Second != 0 || tup.Third != false || tup.Fourth != 0.0 || tup.Fifth != 0 {
		t.Errorf("zero Tuple5 = %v", tup)
	}
}

func BenchmarkNewTuple3(b *testing.B) {
	for range b.N {
		_ = NewTuple3("a", 1, true)
	}
}

func BenchmarkNewTuple5(b *testing.B) {
	for range b.N {
		_ = NewTuple5("a", 1, true, 3.14, byte(0))
	}
}

func FuzzTuple3Unpack(f *testing.F) {
	f.Add("hello", 42, true)
	f.Add("", 0, false)
	f.Fuzz(func(t *testing.T, a string, b int, c bool) {
		tup := NewTuple3(a, b, c)
		ga, gb, gc := tup.Unpack()
		if ga != a || gb != b || gc != c {
			t.Errorf("roundtrip failed for %v, %v, %v", a, b, c)
		}
	})
}
