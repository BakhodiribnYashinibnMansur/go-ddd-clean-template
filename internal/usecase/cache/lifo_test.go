package cache_test

import (
	"testing"

	"gct/internal/usecase/cache"

	"github.com/stretchr/testify/assert"
)

type LIFOTestCase struct {
	Name        string
	Capacity    int
	Operations  []LIFOOperation
	ExpectedLen int
}

type LIFOOperation struct {
	Type       string // "set", "get", "remove", "purge"
	Key        string
	Value      any
	Expected   any
	ExpectedOK bool
}

func TestLIFOCache_TableDriven(t *testing.T) {
	testCases := []LIFOTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},
				{"set", "3", "three", nil, false}, // Should evict "2" (most recently added)
				{"get", "2", nil, nil, false},     // "2" should be evicted
				{"get", "1", nil, "one", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"remove", "1", nil, nil, false},
				{"get", "1", nil, nil, false}, // "1" should be gone
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Purge operation",
			Capacity: 2,
			Operations: []LIFOOperation{
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
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "1", "one_updated", nil, false}, // Update existing key
				{"set", "3", "three", nil, false},       // Should still evict "2" (most recently added)
				{"get", "2", nil, nil, false},           // "2" should be evicted
				{"get", "1", nil, "one_updated", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Zero capacity",
			Capacity: 0,
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"get", "1", nil, nil, false}, // Should not be stored
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false}, // Should evict "1"
				{"get", "1", nil, nil, false},   // "1" should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 1,
		},
		{
			Name:     "Multiple removes",
			Capacity: 3,
			Operations: []LIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "3", "three", nil, false},
				{"remove", "2", nil, nil, false}, // Remove middle item
				{"get", "1", nil, "one", true},
				{"get", "2", nil, nil, false}, // "2" should be gone
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewLIFOCache(tc.Capacity)

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
