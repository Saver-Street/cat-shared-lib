package ratelimit

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkAllow_UniqueKeys(b *testing.B) {
	l := New(Config{
		Rate:            1000,
		Burst:           2000,
		CleanupInterval: time.Minute,
		MaxIdleTime:     time.Minute,
	})
	defer l.Stop()

	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("10.0.%d.%d", i/256, i%256)
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		l.Allow(keys[i%len(keys)])
		i++
	}
}

func BenchmarkAllowN(b *testing.B) {
	l := New(Config{
		Rate:            1000,
		Burst:           2000,
		CleanupInterval: time.Minute,
		MaxIdleTime:     time.Minute,
	})
	defer l.Stop()

	b.ResetTimer()
	for b.Loop() {
		l.AllowN("api-key", 5)
	}
}

func BenchmarkLen(b *testing.B) {
	l := New(Config{
		Rate:            1000,
		Burst:           2000,
		CleanupInterval: time.Minute,
		MaxIdleTime:     time.Minute,
	})
	defer l.Stop()

	for i := 0; i < 100; i++ {
		l.Allow(fmt.Sprintf("key-%d", i))
	}

	b.ResetTimer()
	for b.Loop() {
		l.Len()
	}
}
