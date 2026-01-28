package cache_test

import (
	"testing"

	"gct/pkg/cache"
	"github.com/stretchr/testify/assert"
)

type LRUTestCase struct {
	Name        string
	Capacity    int
	Operations  []LRUOperation
	ExpectedLen int
	ExpectError bool
}

type LRUOperation struct {
	Type       string // "set", "get", "remove", "purge"
	Key        string
	Value      any
	Expected   any
	ExpectedOK bool
}

func TestLRUCache_TableDriven(t *testing.T) {
	testCases := []LRUTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []LRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},
				{"set", "3", "three", nil, false}, // Should evict 2 (LRU)
				{"get", "2", nil, nil, false},     // 2 should be evicted
				{"get", "1", nil, "one", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []LRUOperation{
				{"set", "1", "one", nil, false},
				{"remove", "1", nil, nil, false},
				{"get", "1", nil, nil, false}, // 1 should be gone
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Purge operation",
			Capacity: 2,
			Operations: []LRUOperation{
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
			Operations: []LRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "1", "one_updated", nil, false}, // Update 1 (makes it MRU)
				{"set", "3", "three", nil, false},       // Should evict 2 (LRU)
				{"get", "2", nil, nil, false},           // 2 should be evicted
				{"get", "1", nil, "one_updated", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []LRUOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false}, // Should evict 1
				{"get", "1", nil, nil, false},   // 1 should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 1,
		},
		{
			Name:        "Zero capacity",
			Capacity:    0,
			Operations:  []LRUOperation{},
			ExpectedLen: 0,
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			c := cache.NewLRUCache(tc.Capacity)

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
