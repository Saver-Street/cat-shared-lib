package randutil_test

import (
	"regexp"
	"slices"
	"sort"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/randutil"
)

// --- Int tests ---

func TestIntInRange(t *testing.T) {
	for range 1000 {
		v := randutil.Int(5, 10)
		if v < 5 || v >= 10 {
			t.Fatalf("Int(5,10) = %d; want [5,10)", v)
		}
	}
}

func TestIntSingleValue(t *testing.T) {
	for range 100 {
		v := randutil.Int(7, 8)
		if v != 7 {
			t.Fatalf("Int(7,8) = %d; want 7", v)
		}
	}
}

func TestIntNegativeRange(t *testing.T) {
	for range 1000 {
		v := randutil.Int(-10, -5)
		if v < -10 || v >= -5 {
			t.Fatalf("Int(-10,-5) = %d; want [-10,-5)", v)
		}
	}
}

func TestIntPanicsMinGeMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for min >= max")
		}
	}()
	randutil.Int(5, 5)
}

// --- Float64 tests ---

func TestFloat64InRange(t *testing.T) {
	for range 1000 {
		v := randutil.Float64(1.0, 2.0)
		if v < 1.0 || v >= 2.0 {
			t.Fatalf("Float64(1,2) = %f; want [1,2)", v)
		}
	}
}

func TestFloat64NegativeRange(t *testing.T) {
	for range 1000 {
		v := randutil.Float64(-5.0, -1.0)
		if v < -5.0 || v >= -1.0 {
			t.Fatalf("Float64(-5,-1) = %f; want [-5,-1)", v)
		}
	}
}

// --- Bool tests ---

func TestBoolReturnsBothValues(t *testing.T) {
	seenTrue, seenFalse := false, false
	for range 1000 {
		if randutil.Bool() {
			seenTrue = true
		} else {
			seenFalse = true
		}
		if seenTrue && seenFalse {
			return
		}
	}
	t.Fatal("Bool() did not produce both true and false in 1000 tries")
}

// --- Pick tests ---

func TestPickReturnsElement(t *testing.T) {
	items := []string{"a", "b", "c", "d"}
	for range 100 {
		v := randutil.Pick(items)
		if !slices.Contains(items, v) {
			t.Fatalf("Pick returned %q not in items", v)
		}
	}
}

func TestPickSingleElement(t *testing.T) {
	for range 100 {
		v := randutil.Pick([]int{42})
		if v != 42 {
			t.Fatalf("Pick([42]) = %d; want 42", v)
		}
	}
}

func TestPickPanicsEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty slice")
		}
	}()
	randutil.Pick([]int{})
}

// --- Shuffle tests ---

func TestShuffleSameLength(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	result := randutil.Shuffle(items)
	if len(result) != len(items) {
		t.Fatalf("Shuffle length = %d; want %d", len(result), len(items))
	}
}

func TestShuffleSameElements(t *testing.T) {
	items := []int{3, 1, 4, 1, 5, 9}
	result := randutil.Shuffle(items)
	sortedOrig := make([]int, len(items))
	copy(sortedOrig, items)
	sort.Ints(sortedOrig)
	sort.Ints(result)
	for i := range sortedOrig {
		if sortedOrig[i] != result[i] {
			t.Fatalf("Shuffle changed elements")
		}
	}
}

func TestShuffleDoesNotMutateOriginal(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	orig := make([]int, len(items))
	copy(orig, items)
	randutil.Shuffle(items)
	for i := range orig {
		if orig[i] != items[i] {
			t.Fatal("Shuffle mutated original slice")
		}
	}
}

// --- Sample tests ---

func TestSampleCorrectLength(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := randutil.Sample(items, 3)
	if len(result) != 3 {
		t.Fatalf("Sample length = %d; want 3", len(result))
	}
}

func TestSampleElementsFromOriginal(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	result := randutil.Sample(items, 2)
	for _, v := range result {
		if !slices.Contains(items, v) {
			t.Fatalf("Sample returned %q not in items", v)
		}
	}
}

func TestSampleNoDuplicates(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := randutil.Sample(items, 5)
	seen := make(map[int]bool)
	for _, v := range result {
		if seen[v] {
			t.Fatalf("Sample returned duplicate %d", v)
		}
		seen[v] = true
	}
}

func TestSamplePanicsNExceedsLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for n > len(items)")
		}
	}()
	randutil.Sample([]int{1, 2}, 5)
}

// --- String tests ---

func TestStringCorrectLength(t *testing.T) {
	s := randutil.String(20, "abc")
	if len(s) != 20 {
		t.Fatalf("String length = %d; want 20", len(s))
	}
}

func TestStringCharsFromAlphabet(t *testing.T) {
	alphabet := "xyz"
	s := randutil.String(100, alphabet)
	for _, c := range s {
		if c != 'x' && c != 'y' && c != 'z' {
			t.Fatalf("String contained %c not in alphabet %q", c, alphabet)
		}
	}
}

func TestStringEmptyLength(t *testing.T) {
	s := randutil.String(0, "abc")
	if s != "" {
		t.Fatalf("String(0, ...) = %q; want empty", s)
	}
}

// --- Hex tests ---

func TestHexCorrectLength(t *testing.T) {
	s := randutil.Hex(32)
	if len(s) != 32 {
		t.Fatalf("Hex length = %d; want 32", len(s))
	}
}

func TestHexValidChars(t *testing.T) {
	s := randutil.Hex(100)
	if !regexp.MustCompile(`^[0-9a-f]+$`).MatchString(s) {
		t.Fatalf("Hex contained invalid chars: %s", s)
	}
}

// --- Alpha tests ---

func TestAlphaCorrectLength(t *testing.T) {
	s := randutil.Alpha(15)
	if len(s) != 15 {
		t.Fatalf("Alpha length = %d; want 15", len(s))
	}
}

func TestAlphaOnlyLetters(t *testing.T) {
	s := randutil.Alpha(200)
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(s) {
		t.Fatalf("Alpha contained non-letter chars: %s", s)
	}
}

// --- AlphaNum tests ---

func TestAlphaNumCorrectLength(t *testing.T) {
	s := randutil.AlphaNum(25)
	if len(s) != 25 {
		t.Fatalf("AlphaNum length = %d; want 25", len(s))
	}
}

func TestAlphaNumValidChars(t *testing.T) {
	s := randutil.AlphaNum(200)
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s) {
		t.Fatalf("AlphaNum contained invalid chars: %s", s)
	}
}

// --- WeightedPick tests ---

func TestWeightedPickReturnsFromItems(t *testing.T) {
	items := []string{"a", "b", "c"}
	weights := []float64{1, 2, 3}
	for range 100 {
		v := randutil.WeightedPick(items, weights)
		if !slices.Contains(items, v) {
			t.Fatalf("WeightedPick returned %q not in items", v)
		}
	}
}

func TestWeightedPickSingleWeight(t *testing.T) {
	items := []string{"only"}
	weights := []float64{1.0}
	for range 100 {
		v := randutil.WeightedPick(items, weights)
		if v != "only" {
			t.Fatalf("WeightedPick = %q; want only", v)
		}
	}
}

func TestWeightedPickPanicsMismatchedLengths(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for mismatched lengths")
		}
	}()
	randutil.WeightedPick([]string{"a", "b"}, []float64{1.0})
}

func TestWeightedPickPanicsZeroWeights(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero weights")
		}
	}()
	randutil.WeightedPick([]string{"a", "b"}, []float64{0, 0})
}

// --- Benchmarks ---

func BenchmarkPick(b *testing.B) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for b.Loop() {
		randutil.Pick(items)
	}
}

func BenchmarkShuffle(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		randutil.Shuffle(items)
	}
}

func BenchmarkString(b *testing.B) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	for b.Loop() {
		randutil.String(32, alphabet)
	}
}
