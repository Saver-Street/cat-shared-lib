package cache

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkCache_Set(b *testing.B) {
	c := New[string, string](Config{MaxEntries: 10000, DefaultTTL: time.Minute})
	defer c.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key-%d", i%1000), "value")
	}
}

func BenchmarkCache_Get_Hit(b *testing.B) {
	c := New[string, string](Config{MaxEntries: 10000, DefaultTTL: time.Minute})
	defer c.Stop()

	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key-%d", i%1000))
	}
}

func BenchmarkCache_Get_Miss(b *testing.B) {
	c := New[string, string](Config{MaxEntries: 1000, DefaultTTL: time.Minute})
	defer c.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("missing-%d", i))
	}
}

func BenchmarkCache_SetWithTTL(b *testing.B) {
	c := New[string, int](Config{MaxEntries: 10000})
	defer c.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetWithTTL(fmt.Sprintf("key-%d", i%500), i, 30*time.Second)
	}
}

func BenchmarkCache_Set_Parallel(b *testing.B) {
	c := New[string, string](Config{MaxEntries: 10000, DefaultTTL: time.Minute})
	defer c.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key-%d", i%1000), "value")
			i++
		}
	})
}

func BenchmarkCache_Get_Parallel(b *testing.B) {
	c := New[string, string](Config{MaxEntries: 10000, DefaultTTL: time.Minute})
	defer c.Stop()

	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key-%d", i), "value")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key-%d", i%1000))
			i++
		}
	})
}
