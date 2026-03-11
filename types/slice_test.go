package types

import (
	"testing"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		size  int
		want  [][]int
	}{
		{"even split", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"uneven split", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"size larger", []int{1, 2}, 5, [][]int{{1, 2}}},
		{"size one", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
		{"empty", []int{}, 2, nil},
		{"nil", nil, 2, nil},
		{"zero size", []int{1, 2}, 0, nil},
		{"negative size", []int{1}, -1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Chunk(tt.items, tt.size)
			if len(got) != len(tt.want) {
				t.Fatalf("Chunk len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if len(got[i]) != len(tt.want[i]) {
					t.Fatalf("chunk[%d] len = %d, want %d", i, len(got[i]), len(tt.want[i]))
				}
				for j := range got[i] {
					if got[i][j] != tt.want[i][j] {
						t.Fatalf("chunk[%d][%d] = %d, want %d", i, j, got[i][j], tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name  string
		lists [][]int
		want  []int
	}{
		{"normal", [][]int{{1, 2}, {3, 4}, {5}}, []int{1, 2, 3, 4, 5}},
		{"empty inner", [][]int{{}, {1}, {}}, []int{1}},
		{"nil", nil, []int{}},
		{"all empty", [][]int{{}, {}}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.lists)
			if len(got) != len(tt.want) {
				t.Fatalf("Flatten len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Flatten[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	items := []string{"apple", "avocado", "banana", "cherry", "blueberry"}
	got := GroupBy(items, func(s string) byte { return s[0] })

	if len(got['a']) != 2 {
		t.Fatalf("group 'a' len = %d, want 2", len(got['a']))
	}
	if len(got['b']) != 2 {
		t.Fatalf("group 'b' len = %d, want 2", len(got['b']))
	}
	if len(got['c']) != 1 {
		t.Fatalf("group 'c' len = %d, want 1", len(got['c']))
	}
}

func TestGroupBy_Empty(t *testing.T) {
	got := GroupBy([]int{}, func(n int) int { return n })
	if len(got) != 0 {
		t.Fatalf("GroupBy empty = %d groups, want 0", len(got))
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		want  []int
	}{
		{"normal", []int{1, 2, 3}, []int{3, 2, 1}},
		{"single", []int{1}, []int{1}},
		{"empty", []int{}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reverse(tt.items)
			if len(got) != len(tt.want) {
				t.Fatalf("Reverse len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Reverse[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestZip(t *testing.T) {
	a := []int{1, 2, 3}
	b := []string{"a", "b", "c"}
	got := Zip(a, b)

	if len(got) != 3 {
		t.Fatalf("Zip len = %d, want 3", len(got))
	}
	if got[0].First != 1 || got[0].Second != "a" {
		t.Fatalf("Zip[0] = (%d, %q)", got[0].First, got[0].Second)
	}
}

func TestZip_UnequalLengths(t *testing.T) {
	a := []int{1, 2}
	b := []string{"a", "b", "c"}
	got := Zip(a, b)
	if len(got) != 2 {
		t.Fatalf("Zip len = %d, want 2", len(got))
	}

	// Test b shorter
	got2 := Zip([]int{1, 2, 3}, []string{"x"})
	if len(got2) != 1 {
		t.Fatalf("Zip len = %d, want 1", len(got2))
	}
}

func TestZip_Empty(t *testing.T) {
	got := Zip([]int{}, []string{})
	if len(got) != 0 {
		t.Fatalf("Zip empty len = %d", len(got))
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		name   string
		items  []int
		target int
		want   int
	}{
		{"found first", []int{1, 2, 3}, 1, 0},
		{"found middle", []int{1, 2, 3}, 2, 1},
		{"found last", []int{1, 2, 3}, 3, 2},
		{"not found", []int{1, 2, 3}, 4, -1},
		{"empty", []int{}, 1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Index(tt.items, tt.target)
			if got != tt.want {
				t.Fatalf("Index = %d, want %d", got, tt.want)
			}
		})
	}
}

func BenchmarkChunk(b *testing.B) {
	items := make([]int, 1000)
	for b.Loop() {
		Chunk(items, 10)
	}
}

func BenchmarkFlatten(b *testing.B) {
	lists := make([][]int, 100)
	for i := range lists {
		lists[i] = make([]int, 10)
	}
	for b.Loop() {
		Flatten(lists)
	}
}

func BenchmarkGroupBy(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		GroupBy(items, func(n int) int { return n % 10 })
	}
}

func FuzzChunk(f *testing.F) {
	f.Add(10, 3)
	f.Add(0, 1)
	f.Add(5, 0)
	f.Add(1, 1)

	f.Fuzz(func(t *testing.T, n, size int) {
		if n < 0 || n > 10000 {
			return
		}
		items := make([]int, n)
		_ = Chunk(items, size)
	})
}
