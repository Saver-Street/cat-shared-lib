package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	l := New(Config{})
	defer l.Stop()

	if l.config.Rate != 10 {
		t.Errorf("expected default Rate 10, got %f", l.config.Rate)
	}
	if l.config.Burst != 20 {
		t.Errorf("expected default Burst 20, got %d", l.config.Burst)
	}
	if l.config.CleanupInterval != 5*time.Minute {
		t.Errorf("expected default CleanupInterval 5m, got %v", l.config.CleanupInterval)
	}
	if l.config.MaxIdleTime != 10*time.Minute {
		t.Errorf("expected default MaxIdleTime 10m, got %v", l.config.MaxIdleTime)
	}
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

	if l.config.Rate != 5 {
		t.Errorf("expected Rate 5, got %f", l.config.Rate)
	}
	if l.config.Burst != 10 {
		t.Errorf("expected Burst 10, got %d", l.config.Burst)
	}
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
	if l.Allow("key1") {
		t.Error("request beyond burst should be denied")
	}
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
	if !l.Allow("key1") {
		t.Error("should be allowed after token refill")
	}
}

func TestAllow_PerKeyIsolation(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 2})
	defer l.Stop()

	l.Allow("a")
	l.Allow("a")
	if l.Allow("a") {
		t.Error("key 'a' should be exhausted")
	}
	if !l.Allow("b") {
		t.Error("key 'b' should be independent and allowed")
	}
}

func TestAllowN_Zero(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 1})
	defer l.Stop()

	if !l.AllowN("k", 0) {
		t.Error("AllowN with n=0 should always return true")
	}
	if !l.AllowN("k", -1) {
		t.Error("AllowN with negative n should always return true")
	}
}

func TestAllowN_LargerThanBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 3})
	defer l.Stop()

	if l.AllowN("k", 4) {
		t.Error("AllowN with n > burst should fail on first call")
	}
}

func TestAllowN_ExactBurst(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	if !l.AllowN("k", 5) {
		t.Error("AllowN with n == burst should succeed on first call")
	}
	if l.Allow("k") {
		t.Error("should be denied after using all burst tokens")
	}
}

func TestAllowN_ExistingKey(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	l.Allow("k")
	l.Allow("k")
	// 3 tokens remaining
	if !l.AllowN("k", 3) {
		t.Error("should allow 3 tokens when 3 remain")
	}
	if l.AllowN("k", 1) {
		t.Error("should deny when no tokens remain")
	}
}

func TestStop_Idempotent(t *testing.T) {
	l := New(Config{})
	l.Stop()
	l.Stop() // should not panic
}

func TestLen(t *testing.T) {
	l := New(Config{Rate: 1, Burst: 5})
	defer l.Stop()

	if l.Len() != 0 {
		t.Error("expected 0 keys initially")
	}
	l.Allow("a")
	l.Allow("b")
	if l.Len() != 2 {
		t.Errorf("expected 2 keys, got %d", l.Len())
	}
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

	if l.Len() != 1 {
		t.Errorf("expected 1 key after cleanup, got %d", l.Len())
	}
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
	if !l.Allow("k") {
		t.Error("should be allowed after 1 token refill")
	}
	if l.Allow("k") {
		t.Error("should be denied again, only had 1 token")
	}
}

func BenchmarkAllow(b *testing.B) {
	l := New(Config{Rate: 1000, Burst: 10000})
	defer l.Stop()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
