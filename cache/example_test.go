package cache_test

import (
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/cache"
)

func ExampleCache_Set() {
	c := cache.New[string, int](cache.Config{DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("counter", 42)
	v, ok := c.Get("counter")
	fmt.Println(v, ok)
	// Output:
	// 42 true
}

func ExampleCache_SetWithTTL() {
	c := cache.New[string, string](cache.Config{DefaultTTL: time.Hour})
	defer c.Stop()

	c.SetWithTTL("session", "abc123", 30*time.Second)
	v, ok := c.Get("session")
	fmt.Println(v, ok)
	// Output:
	// abc123 true
}

func ExampleCache_Delete() {
	c := cache.New[string, string](cache.Config{DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key", "value")
	c.Delete("key")
	_, ok := c.Get("key")
	fmt.Println(ok)
	// Output:
	// false
}
