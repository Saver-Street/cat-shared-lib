package types

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLazyGet(t *testing.T) {
	t.Parallel()
	var calls int32
	l := NewLazy(func() string {
		atomic.AddInt32(&calls, 1)
		return "hello"
	})

	got := l.Get()
	if got != "hello" {
		t.Errorf("Get() = %q; want hello", got)
	}

	// Second call should return cached value
	got = l.Get()
	if got != "hello" {
		t.Errorf("Get() = %q; want hello", got)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("init called %d times; want 1", calls)
	}
}

func TestLazyConcurrent(t *testing.T) {
	t.Parallel()
	var calls int32
	l := NewLazy(func() int {
		atomic.AddInt32(&calls, 1)
		return 42
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := l.Get()
			if v != 42 {
				t.Errorf("Get() = %d; want 42", v)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("init called %d times; want 1", calls)
	}
}

func TestLazyZeroValue(t *testing.T) {
	t.Parallel()
	l := NewLazy(func() int {
		return 0
	})
	if got := l.Get(); got != 0 {
		t.Errorf("Get() = %d; want 0", got)
	}
}

func TestLazyErrSuccess(t *testing.T) {
	t.Parallel()
	var calls int32
	l := NewLazyErr(func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})

	got, err := l.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != "ok" {
		t.Errorf("Get() = %q; want ok", got)
	}

	// Second call cached
	got, err = l.Get()
	if err != nil || got != "ok" {
		t.Errorf("Get() = %q, %v; want ok, nil", got, err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("init called %d times; want 1", calls)
	}
}

func TestLazyErrRetry(t *testing.T) {
	t.Parallel()
	var calls int32
	l := NewLazyErr(func() (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return "", errors.New("not ready")
		}
		return "done", nil
	})

	// First two calls fail
	_, err := l.Get()
	if err == nil {
		t.Error("Get() should fail on first call")
	}
	_, err = l.Get()
	if err == nil {
		t.Error("Get() should fail on second call")
	}

	// Third call succeeds
	got, err := l.Get()
	if err != nil {
		t.Fatalf("Get() error = %v on third call", err)
	}
	if got != "done" {
		t.Errorf("Get() = %q; want done", got)
	}

	// Fourth call should be cached
	got, err = l.Get()
	if err != nil || got != "done" {
		t.Errorf("Get() = %q, %v; want done, nil", got, err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("init called %d times; want 3", calls)
	}
}

func TestLazyErrConcurrent(t *testing.T) {
	t.Parallel()
	var calls int32
	l := NewLazyErr(func() (int, error) {
		atomic.AddInt32(&calls, 1)
		return 99, nil
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := l.Get()
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}
			if v != 99 {
				t.Errorf("Get() = %d; want 99", v)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkLazyGet(b *testing.B) {
	l := NewLazy(func() int { return 42 })
	for range b.N {
		l.Get()
	}
}

func BenchmarkLazyErrGet(b *testing.B) {
	l := NewLazyErr(func() (int, error) { return 42, nil })
	for range b.N {
		_, _ = l.Get()
	}
}

func FuzzLazyString(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("test")
	f.Fuzz(func(t *testing.T, s string) {
		l := NewLazy(func() string { return s })
		if got := l.Get(); got != s {
			t.Errorf("Get() = %q; want %q", got, s)
		}
	})
}
