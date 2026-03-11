package types

import "testing"

func TestBloomFilterAddContains(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(100, 0.01)
	bf.AddString("hello")
	bf.AddString("world")
	if !bf.ContainsString("hello") {
		t.Error("ContainsString(hello) = false")
	}
	if !bf.ContainsString("world") {
		t.Error("ContainsString(world) = false")
	}
}

func TestBloomFilterNoFalseNegatives(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(1000, 0.01)
	items := make([]string, 500)
	for i := range items {
		items[i] = string(rune('a'+i%26)) + string(rune('0'+i%10))
		bf.AddString(items[i])
	}
	for _, item := range items {
		if !bf.ContainsString(item) {
			t.Errorf("false negative for %q", item)
		}
	}
}

func TestBloomFilterBytes(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(100, 0.01)
	data := []byte{0x01, 0x02, 0x03}
	bf.Add(data)
	if !bf.Contains(data) {
		t.Error("Contains(data) = false")
	}
}

func TestBloomFilterEmpty(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(100, 0.01)
	// Probabilistically, a random string should not be contained.
	// With 0 elements, the filter should definitely be empty.
	if bf.ContainsString("not_added") {
		t.Error("ContainsString on empty filter returned true")
	}
}

func TestBloomFilterPanicN(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for n=0")
		}
	}()
	NewBloomFilter(0, 0.01)
}

func TestBloomFilterPanicFPRateZero(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for fpRate=0")
		}
	}()
	NewBloomFilter(100, 0)
}

func TestBloomFilterPanicFPRateOne(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for fpRate=1")
		}
	}()
	NewBloomFilter(100, 1)
}

func TestBloomFilterOptimalK(t *testing.T) {
	t.Parallel()
	// When m/n ratio is very small, k should be at least 1.
	k := optimalK(1, 1000)
	if k < 1 {
		t.Errorf("optimalK(1, 1000) = %d; want >= 1", k)
	}
}

func TestBloomFilterLargeFPRate(t *testing.T) {
	t.Parallel()
	bf := NewBloomFilter(10, 0.99)
	bf.AddString("x")
	if !bf.ContainsString("x") {
		t.Error("ContainsString(x) = false after Add")
	}
}

func BenchmarkBloomFilterAdd(b *testing.B) {
	bf := NewBloomFilter(10000, 0.01)
	data := []byte("benchmark-data")
	for range b.N {
		bf.Add(data)
	}
}

func BenchmarkBloomFilterContains(b *testing.B) {
	bf := NewBloomFilter(10000, 0.01)
	for i := range 1000 {
		bf.Add([]byte{byte(i)})
	}
	data := []byte("test")
	b.ResetTimer()
	for range b.N {
		bf.Contains(data)
	}
}

func FuzzBloomFilter(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte(""))
	f.Add([]byte{0xFF, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		bf := NewBloomFilter(100, 0.01)
		bf.Add(data)
		if !bf.Contains(data) {
			t.Error("false negative after Add")
		}
	})
}
