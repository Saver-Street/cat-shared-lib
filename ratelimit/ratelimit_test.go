package ratelimit

import (
	"sync"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNew_Defaults(t *testing.T) {
	l := New(Config{})
	defer l.Stop()

	testkit.AssertEqual(t, l.config.Rate, float64(10))
	testkit.AssertEqual(t, l.config.Burst, 20)
	testkit.AssertEqual(t, l.config.CleanupInterval, 5*time.Minute)
	testkit.AssertEqual(t, l.config.MaxIdleTime, 10*time.Minute)
}

func TestNew_CustomConfig(t *testing.T) {
	cfg := Config{
		Rate:            5,
		Burst:           10,
		CleanupInterval: time.Minute,
		MaxIdleTime:     2 * time.Minute,
	}
	l := New(cfg)
	defer l.Stop()

	testkit.AssertEqual(t, l.config.Rate, float64(5))
	testkit.AssertEqual(t, l.config.Burst, 10)
}

func TestAllow_WithinBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	for i := 0; i < 5; i++ {
		if !l.Allow("key1") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestAllow_ExceedsBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 3})
	defer l.Stop()

	for i := 0; i < 3; i++ {
		if !l.Allow("key1") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
	testkit.AssertFalse(t, l.Allow("key1"))
}

func TestAllow_TokenRefill(t *testing.T) {
	now := time.Now()
	l := New(Config{Rate: 10, Burst: 5})
	defer l.Stop()
	l.now = func() time.Time { return now }

	// Exhaust all tokens.
	for i := 0; i < 5; i++ {
		l.Allow("key1")
	}
	if l.Allow("key1") {
		t.Fatal("should be denied after exhausting burst")
	}

	// Advance time by 1 second: 10 tokens/sec * 1s = 10 tokens, capped at burst=5.
	now = now.Add(time.Second)
	testkit.AssertTrue(t, l.Allow("key1"))
}

func TestAllow_PerKeyIsolation(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 2})
	defer l.Stop()

	l.Allow("a")
	l.Allow("a")
	testkit.AssertFalse(t, l.Allow("a"))
	testkit.AssertTrue(t, l.Allow("b"))
}

func TestAllowN_Zero(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 1})
	defer l.Stop()

	testkit.AssertTrue(t, l.AllowN("k", 0))
	testkit.AssertTrue(t, l.AllowN("k", -1))
}

func TestAllowN_LargerThanBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 3})
	defer l.Stop()

	testkit.AssertFalse(t, l.AllowN("k", 4))
}

func TestAllowN_ExactBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	testkit.AssertTrue(t, l.AllowN("k", 5))
	testkit.AssertFalse(t, l.Allow("k"))
}

func TestAllowN_ExistingKey(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	l.Allow("k")
	l.Allow("k")
	// 3 tokens remaining
	testkit.AssertTrue(t, l.AllowN("k", 3))
	testkit.AssertFalse(t, l.AllowN("k", 1))
}

func TestStop_Idempotent(t *testing.T) {
	l := New(Config{CleanupInterval: 10 * time.Millisecond})
	// Let the cleanup ticker fire at least once.
	time.Sleep(50 * time.Millisecond)
	l.Stop()
	l.Stop() // should not panic
}

func TestLen(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	testkit.AssertEqual(t, l.Len(), 0)
	l.Allow("a")
	l.Allow("b")
	testkit.AssertEqual(t, l.Len(), 2)
}

func TestCleanup_RemovesIdleEntries(t *testing.T) {
	now := time.Now()
	l := New(Config{Rate: 1, Burst: 5, MaxIdleTime: time.Minute})
	defer l.Stop()
	l.now = func() time.Time { return now }

	l.Allow("stale")
	l.Allow("fresh")

	// Advance time past MaxIdleTime for "stale" only.
	now = now.Add(2 * time.Minute)
	l.Allow("fresh") // refresh "fresh"

	l.cleanup()

	testkit.AssertEqual(t, l.Len(), 1)
}

func TestConcurrent_Allow(t *testing.T) {
	l := New(Config{Rate: 100, Burst: 1000})
	defer l.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				l.Allow("concurrent-key")
			}
		}()
	}
	wg.Wait()
}

func TestAllow_DeniedDoesNotConsume(t *testing.T) {
	now := time.Now()
	l := New(Config{Rate: 1, Burst: 2})
	defer l.Stop()
	l.now = func() time.Time { return now }

	l.Allow("k")
	l.Allow("k")
	// All tokens consumed, next should fail.
	if l.Allow("k") {
		t.Fatal("should be denied")
	}

	// Advance time enough for exactly 1 token.
	now = now.Add(time.Second)
	testkit.AssertTrue(t, l.Allow("k"))
	testkit.AssertFalse(t, l.Allow("k"))
}

func BenchmarkAllow(b *testing.B) {
	l := New(Config{Rate: 1000, Burst: 10000})
	defer l.Stop()
	b.ResetTimer()
	for b.Loop() {
		l.Allow("bench-key")
	}
}

func BenchmarkAllow_Parallel(b *testing.B) {
	l := New(Config{Rate: 1000, Burst: 10000})
	defer l.Stop()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Allow("bench-key")
		}
	})
}
