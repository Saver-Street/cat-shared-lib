package ratelimit

import (
	"testing"
	"time"
)

func FuzzAllow(f *testing.F) {
	f.Add("192.168.1.1")
	f.Add("")
	f.Add("user:admin")
	f.Add("unicode-キー")
	f.Add("very-long-key-" + string(make([]byte, 200)))
	f.Fuzz(func(t *testing.T, key string) {
		l := New(Config{
			Rate:            100,
			Burst:           200,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		// First call should always be allowed with a high burst.
		if !l.Allow(key) {
			t.Error("first Allow returned false with high burst")
		}
	})
}

func FuzzAllowN(f *testing.F) {
	f.Add("key", 1)
	f.Add("", 0)
	f.Add("k", -1)
	f.Add("key", 100)
	f.Fuzz(func(t *testing.T, key string, n int) {
		l := New(Config{
			Rate:            100,
			Burst:           200,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		// Must not panic regardless of n.
		_ = l.AllowN(key, n)
	})
}

func FuzzBurstExhaustion(f *testing.F) {
	f.Add("ip-1", 1, 5)
	f.Add("ip-2", 10, 10)
	f.Add("", 1, 1)
	f.Add("key", 0, 0)
	f.Add("key", -1, -5)
	f.Add("key", 1000, 5000)

	f.Fuzz(func(t *testing.T, key string, burst, requests int) {
		if burst <= 0 {
			burst = 1
		}
		if burst > 1000 {
			burst = 1000
		}
		if requests <= 0 {
			requests = 1
		}
		if requests > 2000 {
			requests = 2000
		}

		l := New(Config{
			Rate:            0.001, // nearly zero refill
			Burst:           burst,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		allowed := 0
		for range requests {
			if l.Allow(key) {
				allowed++
			}
		}
		// Cannot allow more than burst tokens.
		if allowed > burst {
			t.Errorf("allowed %d > burst %d for key %q", allowed, burst, key)
		}
	})
}

func FuzzConcurrentAllow(f *testing.F) {
	f.Add("shared-key", 4)
	f.Add("", 2)
	f.Add("unicode-キー", 8)
	f.Add("key\x00with\x00nulls", 6)

	f.Fuzz(func(t *testing.T, key string, goroutines int) {
		if goroutines <= 0 {
			goroutines = 1
		}
		if goroutines > 32 {
			goroutines = 32
		}

		l := New(Config{
			Rate:            100,
			Burst:           200,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		done := make(chan bool, goroutines)
		for range goroutines {
			go func() {
				done <- l.Allow(key)
			}()
		}
		for range goroutines {
			<-done
		}
		// Must not panic or deadlock (reaching here = success).
	})
}

func FuzzLenTracking(f *testing.F) {
	f.Add("a", "b", "c")
	f.Add("", "", "")
	f.Add("same", "same", "same")
	f.Add("key\x00a", "key\x00b", "日本語")

	f.Fuzz(func(t *testing.T, k1, k2, k3 string) {
		l := New(Config{
			Rate:            100,
			Burst:           200,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		l.Allow(k1)
		l.Allow(k2)
		l.Allow(k3)

		n := l.Len()
		// Count unique keys.
		unique := map[string]bool{k1: true, k2: true, k3: true}
		expected := len(unique)
		if n != expected {
			t.Errorf("Len() = %d, want %d unique keys from %q,%q,%q",
				n, expected, k1, k2, k3)
		}
	})
}

func FuzzAllowNBurstMath(f *testing.F) {
	f.Add("key", 5, 10)
	f.Add("key", 10, 10)
	f.Add("key", 11, 10)
	f.Add("key", 0, 10)
	f.Add("key", -1, 10)
	f.Add("key", 1, 1)

	f.Fuzz(func(t *testing.T, key string, n, burst int) {
		if burst <= 0 {
			burst = 1
		}
		if burst > 10000 {
			burst = 10000
		}

		l := New(Config{
			Rate:            0.001,
			Burst:           burst,
			CleanupInterval: time.Minute,
			MaxIdleTime:     time.Minute,
		})
		defer l.Stop()

		result := l.AllowN(key, n)
		if n <= 0 {
			// Requesting zero or negative tokens always succeeds.
			if !result {
				t.Errorf("AllowN(%q, %d) = false, want true for n <= 0", key, n)
			}
		} else if n > burst {
			// Requesting more than burst capacity must fail.
			if result {
				t.Errorf("AllowN(%q, %d) = true with burst %d, want false", key, n, burst)
			}
		}
	})
}
