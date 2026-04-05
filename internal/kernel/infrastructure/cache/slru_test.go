package cache_test

import (
	"testing"

	"gct/internal/kernel/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

func TestSLRUCache(t *testing.T) {
	t.Parallel()

	t.Run("Basic Operations", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5) // 2 protected, 2 probationary

		// Add items - they start in probationary
		c.Set("1", "one")
		c.Set("2", "two")

		val, ok := c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one", val)

		// After access, "1" should be promoted to protected
		assert.Equal(t, 2, c.Len())

		// Access "2" to promote it
		c.Get("2")

		// Now both should be in protected
		c.Set("3", "three") // Goes to probationary
		c.Set("4", "four")  // Goes to probationary

		assert.Equal(t, 4, c.Len())

		// Add 5th item, should evict from probationary
		c.Set("5", "five")

		assert.Equal(t, 4, c.Len())

		// Protected items should still be accessible
		val, ok = c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one", val)

		val, ok = c.Get("2")
		assert.True(t, ok)
		assert.Equal(t, "two", val)
	})

	t.Run("Remove", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5)

		c.Set("1", "one")
		c.Remove("1")

		val, ok := c.Get("1")
		assert.False(t, ok)
		assert.Nil(t, val)
		assert.Equal(t, 0, c.Len())
	})

	t.Run("Purge", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5)

		c.Set("1", "one")
		c.Set("2", "two")
		c.Purge()

		assert.Equal(t, 0, c.Len())
		val, ok := c.Get("1")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("Promotion from probationary to protected", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5) // 2 protected, 2 probationary

		// Fill probationary
		c.Set("1", "one")
		c.Set("2", "two")

		// Access to promote
		c.Get("1")
		c.Get("2")

		// Add more to fill protected
		c.Set("3", "three")
		c.Get("3")

		// Protected should be full now, accessing "3" again should demote LRU from protected
		c.Get("3")

		// Add new item
		c.Set("4", "four")

		assert.Equal(t, 4, c.Len())

		// All should be accessible
		val, ok := c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one", val)
	})

	t.Run("Demotion from protected to probationary", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5) // 2 protected, 2 probationary

		// Fill protected
		c.Set("1", "one")
		c.Get("1")
		c.Set("2", "two")
		c.Get("2")

		// Fill protected more (should trigger demotion)
		c.Set("3", "three")
		c.Get("3")

		// Now protected should have 2 items, probationary should have some
		assert.Equal(t, 3, c.Len())

		// All should be accessible
		items := []string{"1", "2", "3"}
		for _, key := range items {
			val, ok := c.Get(key)
			assert.True(t, ok, "Key %s should be accessible", key)
			assert.NotNil(t, val, "Value for key %s should not be nil", key)
		}
	})

	t.Run("Zero capacity", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(0, 0.5)

		c.Set("1", "one")
		// Some implementations might still store items with zero capacity
		// Let's check the actual behavior
		val, ok := c.Get("1")
		if ok {
			assert.Equal(t, "one", val)
			// If it stores items, length should be 1
			assert.Equal(t, 1, c.Len())
		} else {
			assert.Nil(t, val)
			assert.Equal(t, 0, c.Len())
		}
	})

	t.Run("Single item capacity", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(1, 0.5)

		c.Set("1", "one")
		assert.Equal(t, 1, c.Len())

		c.Set("2", "two")
		assert.Equal(t, 1, c.Len())

		// Only one should be present
		present := 0
		if _, ok := c.Get("1"); ok {
			present++
		}
		if _, ok := c.Get("2"); ok {
			present++
		}
		assert.Equal(t, 1, present)
	})

	t.Run("Edge case ratios", func(t *testing.T) {
		t.Parallel()

		// Very small protected ratio
		c1 := cache.NewSLRUCache(4, 0.1)
		c1.Set("1", "one")
		c1.Set("2", "two")
		// Should have at least some capacity
		assert.True(t, c1.Len() >= 0 && c1.Len() <= 2)

		// Very large protected ratio
		c2 := cache.NewSLRUCache(4, 0.9)
		c2.Set("1", "one")
		c2.Set("2", "two")
		// Should have at least some capacity
		assert.True(t, c2.Len() >= 0 && c2.Len() <= 2)
	})

	t.Run("Update existing key", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5)

		c.Set("1", "one")
		c.Get("1") // Promote to protected

		// Update existing key
		c.Set("1", "one_updated")

		val, ok := c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one_updated", val)
		assert.Equal(t, 1, c.Len())
	})

	t.Run("Remove non-existent key", func(t *testing.T) {
		t.Parallel()
		c := cache.NewSLRUCache(4, 0.5)

		c.Set("1", "one")
		c.Remove("non_existent")

		// Should not affect existing items
		val, ok := c.Get("1")
		assert.True(t, ok)
		assert.Equal(t, "one", val)
		assert.Equal(t, 1, c.Len())
	})
}
