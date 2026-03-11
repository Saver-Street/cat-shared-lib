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
