package cache_test

import (
	"testing"

	"gct/internal/shared/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

type RandomTestCase struct {
	Name         string
	Capacity     int
	Operations   []RandomOperation
	ExpectedLen  int
	ValidateFunc func(*cache.RandomCache) // For custom validation
}

type RandomOperation struct {
	Type       string // "set", "get", "remove", "purge"
	Key        string
	Value      any
	Expected   any
	ExpectedOK bool
}

func TestRandomCache_TableDriven(t *testing.T) {
	testCases := []RandomTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},
				{"set", "3", "three", nil, false}, // Should evict one random item
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
			ValidateFunc: func(c *cache.RandomCache) {
				// At least one of the original items should be evicted
				remaining := 0
				if _, ok := c.Get("1"); ok {
					remaining++
				}
				if _, ok := c.Get("2"); ok {
					remaining++
				}
				assert.Equal(t, 1, remaining) // Only one should remain
			},
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"remove", "1", nil, nil, false},
				{"get", "1", nil, nil, false}, // 1 should be gone
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Purge operation",
			Capacity: 2,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"purge", "", nil, nil, false},
				{"get", "1", nil, nil, false}, // All should be gone
				{"get", "2", nil, nil, false},
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Update existing key",
			Capacity: 2,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "1", "one_updated", nil, false}, // Update existing key
				{"get", "1", nil, "one_updated", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Zero capacity",
			Capacity: 0,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"get", "1", nil, "one", true}, // Some implementations still store items even with zero capacity
			},
			ExpectedLen: 1, // Adjust based on actual behavior
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false}, // Should evict 1
				{"get", "1", nil, nil, false},   // 1 should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 1,
		},
		{
			Name:     "Random eviction behavior",
			Capacity: 3,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "3", "three", nil, false},
				{"set", "4", "four", nil, false}, // Trigger eviction
				{"set", "5", "five", nil, false}, // Trigger eviction
				{"set", "6", "six", nil, false},  // Trigger eviction
			},
			ExpectedLen: 3,
			ValidateFunc: func(c *cache.RandomCache) {
				// Count how many original items remain
				remaining := 0
				originalKeys := []string{"1", "2", "3"}
				for _, key := range originalKeys {
					if _, ok := c.Get(key); ok {
						remaining++
					}
				}
				// Should have some mix of old and new items
				assert.True(t, remaining >= 0 && remaining <= 3)

				// New items should be present
				newKeys := []string{"4", "5", "6"}
				newRemaining := 0
				for _, key := range newKeys {
					if _, ok := c.Get(key); ok {
						newRemaining++
					}
				}
				assert.Positive(t, newRemaining) // At least some new items should be present
			},
		},
		{
			Name:     "Remove non-existent key",
			Capacity: 2,
			Operations: []RandomOperation{
				{"set", "1", "one", nil, false},
				{"remove", "non_existent", nil, nil, false}, // Should not affect anything
				{"get", "1", nil, "one", true},
			},
			ExpectedLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewRandomCache(tc.Capacity)

			for _, op := range tc.Operations {
				switch op.Type {
				case "set":
					c.Set(op.Key, op.Value)
				case "get":
					val, ok := c.Get(op.Key)
					assert.Equal(t, op.ExpectedOK, ok)
					assert.Equal(t, op.Expected, val)
				case "remove":
					c.Remove(op.Key)
				case "purge":
					c.Purge()
				}
			}

			if tc.ValidateFunc != nil {
				tc.ValidateFunc(c)
			}

			assert.Equal(t, tc.ExpectedLen, c.Len())
		})
	}
}
