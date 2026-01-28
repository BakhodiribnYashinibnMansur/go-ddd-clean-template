package cache_test

import (
	"testing"

	"gct/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestTwoQueueCache(t *testing.T) {
	t.Parallel()

	// 2Q usually has a fraction of capacity for In and Out queues.
	// We need to ensure we trigger eviction.

	t.Run("Basic Operations", func(t *testing.T) {
		t.Parallel()
		// Capacity 2: Q1 (In) cap = 1, Q2 (Am) cap = 1
		c := cache.NewTwoQueueCache(2)

		c.Set("1", "one") // "1" -> Q1

		// Access "1" to promote to Q2
		val, ok := c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one", val)
		// Now "1" is in Q2

		c.Set("2", "two") // "2" -> Q1

		// Now we have "1" in Q2, "2" in Q1. Total 2.
		assert.Equal(t, 2, c.Len())

		c.Set("3", "three") // "3" -> Q1. "2" evicted from Q1 to Qout.

		// In memory: "1" (Q2), "3" (Q1). Total 2.
		assert.Equal(t, 2, c.Len())

		_, ok = c.Get("1")
		assert.True(t, ok)

		_, ok = c.Get("3")
		assert.True(t, ok)

		val, ok = c.Get("2")
		assert.False(t, ok, "2 should have been evicted")
		assert.Nil(t, val)
	})

	t.Run("Remove", func(t *testing.T) {
		t.Parallel()
		c := cache.NewTwoQueueCache(10)
		c.Set("1", "one")
		c.Delete("1")
		val, ok := c.Get("1")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("Purge", func(t *testing.T) {
		t.Parallel()
		c := cache.NewTwoQueueCache(10)
		c.Set("1", "one")
		c.Purge()
		assert.Equal(t, 0, c.Len())
	})
}
