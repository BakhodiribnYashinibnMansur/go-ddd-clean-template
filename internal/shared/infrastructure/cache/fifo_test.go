package cache_test

import (
	"testing"

	"gct/internal/shared/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

type FIFOTestCase struct {
	Name        string
	Capacity    int
	Operations  []FIFOOperation
	ExpectedLen int
}

type FIFOOperation struct {
	Type       string // "set", "get", "remove", "purge"
	Key        string
	Value      any
	Expected   any
	ExpectedOK bool
}

func TestFIFOCache_TableDriven(t *testing.T) {
	testCases := []FIFOTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []FIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"get", "1", nil, "one", true},
				{"set", "3", "three", nil, false}, // Should evict "1"
				{"get", "1", nil, nil, false},     // "1" should be evicted
				{"get", "2", nil, "two", true},
				{"get", "3", nil, "three", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Update doesn't affect order",
			Capacity: 2,
			Operations: []FIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false},
				{"set", "1", "one_updated", nil, false}, // Update "1"
				{"set", "3", "three", nil, false},       // Should still evict "1"
				{"get", "1", nil, nil, false},           // "1" should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []FIFOOperation{
				{"set", "1", "one", nil, false},
				{"remove", "1", nil, nil, false},
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Purge operation",
			Capacity: 2,
			Operations: []FIFOOperation{
				{"set", "1", "one", nil, false},
				{"purge", "", nil, nil, false},
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []FIFOOperation{
				{"set", "1", "one", nil, false},
				{"set", "2", "two", nil, false}, // Should evict "1"
				{"get", "1", nil, nil, false},   // "1" should be evicted
				{"get", "2", nil, "two", true},
			},
			ExpectedLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewFIFOCache(tc.Capacity)

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
