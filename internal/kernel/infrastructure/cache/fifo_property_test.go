package cache_test

import (
	"fmt"
	"testing"

	"gct/internal/kernel/infrastructure/cache"

	"pgregory.net/rapid"
)

func TestFIFOCache_Property_CapacityInvariant(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewFIFOCache(cap)
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

func TestFIFOCache_Property_GetAfterSet(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewFIFOCache(cap)
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

func TestFIFOCache_Property_PurgeClearsAll(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewFIFOCache(cap)
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

func TestFIFOCache_Property_FIFOEviction(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(2, 5).Draw(t, "capacity")
		c := cache.NewFIFOCache(cap)

		// insert exactly cap keys
		keys := make([]string, cap)
		for i := 0; i < cap; i++ {
			keys[i] = fmt.Sprintf("k%d", i)
			c.Set(keys[i], i)
		}

		// access keys[0] — should NOT change eviction order in FIFO
		c.Get(keys[0])

		// insert one more — should evict keys[0] (oldest inserted)
		c.Set("new", 999)
		_, ok := c.Get(keys[0])
		if ok {
			t.Fatalf("FIFO: oldest key %q not evicted despite being accessed", keys[0])
		}
	})
}

func TestFIFOCache_Property_AccessInsensitive(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(2, 5).Draw(t, "capacity")
		c := cache.NewFIFOCache(cap)

		keys := make([]string, cap)
		for i := 0; i < cap; i++ {
			keys[i] = fmt.Sprintf("k%d", i)
			c.Set(keys[i], i)
		}

		// access all keys multiple times — should not affect eviction order
		nAccess := rapid.IntRange(1, 20).Draw(t, "nAccess")
		for i := 0; i < nAccess; i++ {
			idx := rapid.IntRange(0, cap-1).Draw(t, fmt.Sprintf("idx%d", i))
			c.Get(keys[idx])
		}

		// insert one more — first key should be evicted regardless of access
		c.Set("new", 999)
		_, ok := c.Get(keys[0])
		if ok {
			t.Fatalf("FIFO: oldest key %q not evicted after repeated access", keys[0])
		}
	})
}
