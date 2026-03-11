package types

import (
	"bytes"
	"sync"
	"testing"
)

func TestPoolGetPut(t *testing.T) {
	t.Parallel()
	p := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	buf := p.Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}
	buf.WriteString("hello")
	if buf.String() != "hello" {
		t.Errorf("buffer = %q; want hello", buf.String())
	}

	buf.Reset()
	p.Put(buf)

	buf2 := p.Get()
	if buf2 == nil {
		t.Fatal("Get() returned nil after Put")
	}
}

func TestPoolCreatesNew(t *testing.T) {
	t.Parallel()
	var created int
	p := NewPool(func() int {
		created++
		return created
	})

	v1 := p.Get()
	if v1 != 1 {
		t.Errorf("Get() = %d; want 1", v1)
	}
	v2 := p.Get()
	if v2 != 2 {
		t.Errorf("Get() = %d; want 2", v2)
	}
}

func TestPoolReuse(t *testing.T) {
	t.Parallel()
	p := NewPool(func() []byte {
		return make([]byte, 0, 1024)
	})

	s := p.Get()
	s = append(s, "test"...)
	p.Put(s[:0])

	// May or may not get the same slice back depending on GC
	s2 := p.Get()
	if s2 == nil {
		t.Error("Get() returned nil")
	}
}

func TestPoolConcurrent(t *testing.T) {
	t.Parallel()
	p := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := p.Get()
			buf.WriteString("data")
			buf.Reset()
			p.Put(buf)
		}()
	}
	wg.Wait()
}

func TestPoolValueType(t *testing.T) {
	t.Parallel()
	p := NewPool(func() int { return 42 })
	v := p.Get()
	if v != 42 {
		t.Errorf("Get() = %d; want 42", v)
	}
	p.Put(99)
	// May return 99 or 42 (GC can clear pool)
	v = p.Get()
	if v != 42 && v != 99 {
		t.Errorf("Get() = %d; want 42 or 99", v)
	}
}

func BenchmarkPoolGetPut(b *testing.B) {
	p := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})
	for range b.N {
		buf := p.Get()
		buf.Reset()
		p.Put(buf)
	}
}

func BenchmarkPoolConcurrent(b *testing.B) {
	p := NewPool(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			buf.Reset()
			p.Put(buf)
		}
	})
}

func FuzzPoolInt(f *testing.F) {
	f.Add(0)
	f.Add(42)
	f.Add(-1)
	f.Fuzz(func(t *testing.T, n int) {
		p := NewPool(func() int { return n })
		v := p.Get()
		if v != n {
			t.Errorf("Get() = %d; want %d", v, n)
		}
		p.Put(v)
	})
}
