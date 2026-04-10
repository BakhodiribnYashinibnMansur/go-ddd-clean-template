package cache_test

import (
	"fmt"
	"testing"

	"gct/internal/kernel/infrastructure/cache"

	"pgregory.net/rapid"
)

func TestLFUCache_Property_CapacityInvariant(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLFUCache(cap)
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

func TestLFUCache_Property_GetAfterSet(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLFUCache(cap)
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

func TestLFUCache_Property_PurgeClearsAll(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(1, 10).Draw(t, "capacity")
		c := cache.NewLFUCache(cap)
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

func TestLFUCache_Property_FrequencyEviction(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		cap := rapid.IntRange(2, 5).Draw(t, "capacity")
		c := cache.NewLFUCache(cap)

		// insert exactly cap keys
		keys := make([]string, cap)
		for i := 0; i < cap; i++ {
			keys[i] = fmt.Sprintf("k%d", i)
			c.Set(keys[i], i)
		}

		// access all keys except keys[0] multiple times to boost their frequency
		nAccess := rapid.IntRange(1, 5).Draw(t, "nAccess")
		for i := 1; i < cap; i++ {
			for j := 0; j < nAccess; j++ {
				c.Get(keys[i])
			}
		}

		// insert one more — keys[0] has lowest frequency and should be evicted
		c.Set("new", 999)
		_, ok := c.Get(keys[0])
		if ok {
			t.Fatalf("LFU: least frequent key %q not evicted", keys[0])
		}
	})
}

func TestLFUCache_Property_SetUpdateIdempotentLen(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		c := cache.NewLFUCache(5)
		key := rapid.SampledFrom([]string{"a", "b", "c"}).Draw(t, "key")

		c.Set(key, 1)
		lenBefore := c.Len()
		c.Set(key, 2)
		if c.Len() != lenBefore {
			t.Fatalf("Len changed from %d to %d after updating same key", lenBefore, c.Len())
		}
	})
}
