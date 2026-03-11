package cache

import (
	"testing"
	"time"
)

func FuzzSetGet(f *testing.F) {
	f.Add("key1", "value1")
	f.Add("", "empty-key")
	f.Add("unicode-🔑", "unicode-💎")
	f.Add("key with spaces", "value with spaces")
	f.Fuzz(func(t *testing.T, key, value string) {
		c := New[string, string](Config{
			MaxEntries:      100,
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Minute,
		})
		defer c.Stop()

		c.Set(key, value)
		got, ok := c.Get(key)
		if !ok {
			t.Error("Get returned false after Set")
		}
		if got != value {
			t.Errorf("Get = %q, want %q", got, value)
		}
	})
}

func FuzzSetWithTTL(f *testing.F) {
	f.Add("key", "val", int64(1000))
	f.Add("", "", int64(0))
	f.Add("k", "v", int64(-1))
	f.Fuzz(func(t *testing.T, key, value string, ttlNs int64) {
		c := New[string, string](Config{
			MaxEntries:      100,
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Minute,
		})
		defer c.Stop()

		ttl := time.Duration(ttlNs)
		// Must not panic regardless of TTL value.
		c.SetWithTTL(key, value, ttl)

		// If TTL is large enough, value should be retrievable immediately.
		if ttl >= time.Second {
			got, ok := c.Get(key)
			if !ok {
				t.Error("Get returned false after SetWithTTL with large TTL")
			}
			if got != value {
				t.Errorf("Get = %q, want %q", got, value)
			}
		}
	})
}

func FuzzDeleteLen(f *testing.F) {
	f.Add("a", "b", "c")
	f.Add("", "", "")
	f.Fuzz(func(t *testing.T, k1, k2, k3 string) {
		c := New[string, string](Config{
			MaxEntries:      100,
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Minute,
		})
		defer c.Stop()

		c.Set(k1, "v1")
		c.Set(k2, "v2")
		c.Set(k3, "v3")

		c.Delete(k1)

		if _, ok := c.Get(k1); ok && k1 != k2 && k1 != k3 {
			t.Error("Get returned true after Delete")
		}

		c.Clear()
		if c.Len() != 0 {
			t.Errorf("Len after Clear = %d, want 0", c.Len())
		}
	})
}
