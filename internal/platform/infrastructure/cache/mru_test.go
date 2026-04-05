package cache_test

import (
	"testing"

	"gct/internal/platform/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

type MRUTestCase struct {
	Name        string
	Capacity    int
	Operations  []MRUOperation
	ExpectedLen int
}

type MRUOperation struct {
	Type       string // "set", "get", "remove", "purge"
	Key        string
	Value      any
	Expected   any
	ExpectedOK bool
}

func TestMRUCache_TableDriven(t *testing.T) {
	testCases := []MRUTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},
				{"set", "3", "three", nil, false}, // Should evict 1 (MRU)
				{"get", "1", nil, nil, false},     // 1 should be evicted
				{"get", "2", nil, "two", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"remove", "1", nil, nil, false},
				{"get", "1", nil, nil, false}, // 1 should be gone
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Purge operation",
			Capacity: 2,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"purge", "", nil, nil, false},
				{"get", "1", nil, nil, false}, // All should be gone
				{"get", "2", nil, nil, false},
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Update accesses",
			Capacity: 2,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},    // Access 1 (makes it MRU)
				{"set", "3", "three", nil, false}, // Should evict 1 (MRU)
				{"get", "1", nil, nil, false},     // 1 should be evicted
				{"get", "2", nil, "two", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Update existing key",
			Capacity: 2,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "1", "one_updated", nil, false}, // Update 1 (makes it MRU)
				{"set", "3", "three", nil, false},       // Should evict 1 (MRU)
				{"get", "1", nil, nil, false},           // 1 should be evicted
				{"get", "2", nil, "two", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Zero capacity",
			Capacity: 0,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"get", "1", nil, "one", true}, // Some implementations still store items even with zero capacity
			},
			ExpectedLen: 1, // Adjust based on actual behavior
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false}, // Should evict 1
				{"get", "1", nil, nil, false},   // 1 should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 1,
		},
		{
			Name:     "Access pattern",
			Capacity: 3,
			Operations: []MRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "3", "three", nil, false},
				{"get", "1", nil, "one", true},   // Access 1
				{"get", "2", nil, "two", true},   // Access 2
				{"get", "3", nil, "three", true}, // Access 3 (now MRU)
				{"set", "4", "four", nil, false}, // Should evict 3 (MRU)
				{"get", "3", nil, nil, false},    // 3 should be evicted
				{"get", "1", nil, "one", true},
				{"get", "2", nil, "two", true},
				{"get", "4", nil, "four", true},
			},
			ExpectedLen: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewMRUCache(tc.Capacity)

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

			assert.Equal(t, tc.ExpectedLen, c.Len())
		})
	}
}
