package cache_test

import (
	"fmt"
	"testing"

	"gct/internal/kernel/infrastructure/cache"

	"pgregory.net/rapid"
)

func TestLRUCache_Property_CapacityInvariant(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLRUCache(cap)
		nOps := rapid.IntRange(1, 100).Draw(t, "nOps")

		for i := 0; i < nOps; i++ {
			op := rapid.SampledFrom([]string{"set", "get", "remove"}).Draw(t, fmt.Sprintf("op%d", i))
			key := rapid.SampledFrom([]string{"a", "b", "c", "d", "e", "f"}).Draw(t, fmt.Sprintf("key%d", i))
			switch op {
			case "set":
				val := rapid.Int().Draw(t, fmt.Sprintf("val%d", i))
				c.Set(key, val)
			case "get":
				c.Get(key)
			case "remove":
				c.Remove(key)
			}
			if c.Len() > cap {
				t.Fatalf("Len() = %d exceeds capacity %d", c.Len(), cap)
			}
		}
	})
}

func TestLRUCache_Property_GetAfterSet(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLRUCache(cap)
		key := rapid.SampledFrom([]string{"a", "b", "c", "d", "e", "f"}).Draw(t, "key")
		val := rapid.Int().Draw(t, "val")

		c.Set(key, val)
		got, ok := c.Get(key)
		if !ok {
			t.Fatalf("Get(%q) returned false after Set", key)
		}
		if got != val {
			t.Fatalf("Get(%q) = %v, want %v", key, got, val)
		}
	})
}

func TestLRUCache_Property_RemoveWorks(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		c := cache.NewLRUCache(10)
		key := rapid.SampledFrom([]string{"a", "b", "c"}).Draw(t, "key")

		c.Set(key, 1)
		c.Remove(key)
		_, ok := c.Get(key)
		if ok {
			t.Fatalf("Get(%q) returned true after Remove", key)
		}
	})
}

func TestLRUCache_Property_PurgeClearsAll(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLRUCache(cap)
		nOps := rapid.IntRange(1, 20).Draw(t, "nOps")

		for i := 0; i < nOps; i++ {
			key := rapid.SampledFrom([]string{"a", "b", "c", "d", "e"}).Draw(t, fmt.Sprintf("key%d", i))
			c.Set(key, i)
		}
		c.Purge()
		if c.Len() != 0 {
			t.Fatalf("Len() = %d after Purge, want 0", c.Len())
		}
	})
}

func TestLRUCache_Property_LRUEviction(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(2, 5).Draw(t, "capacity")
		c := cache.NewLRUCache(cap)

		// insert exactly cap keys
		keys := make([]string, cap)
		for i := 0; i < cap; i++ {
			keys[i] = fmt.Sprintf("k%d", i)
			c.Set(keys[i], i)
		}

		// access all except the first key (keys[0] becomes LRU)
		for i := 1; i < cap; i++ {
			c.Get(keys[i])
		}

		// insert one more — should evict keys[0]
		c.Set("new", 999)
		_, ok := c.Get(keys[0])
		if ok {
			t.Fatalf("LRU key %q not evicted after capacity overflow", keys[0])
		}
		// "new" key should be present
		_, ok = c.Get("new")
		if !ok {
			t.Fatal("newly inserted key not found")
		}
	})
}

func TestLRUCache_Property_SetUpdateIdempotentLen(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		c := cache.NewLRUCache(5)
		key := rapid.SampledFrom([]string{"a", "b", "c"}).Draw(t, "key")

		c.Set(key, 1)
		lenBefore := c.Len()
		c.Set(key, 2)
		if c.Len() != lenBefore {
			t.Fatalf("Len changed from %d to %d after updating same key", lenBefore, c.Len())
		}
	})
}
