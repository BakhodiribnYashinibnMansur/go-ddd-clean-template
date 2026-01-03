package cache_test

import (
	"testing"

	"gct/internal/usecase/cache"

	"github.com/stretchr/testify/assert"
)

type LFUTestCase struct {
	Name        string
	Capacity    int
	Operations  []LFUOperation
	ExpectedLen int
}

type LFUOperation struct {
	Type        string // "set", "get", "remove", "purge"
	Key         string
	Value       any
	Expected    any
	ExpectedOK  bool
	AccessCount int // How many times to access this key
}

func TestLFUCache_TableDriven(t *testing.T) {
	testCases := []LFUTestCase{
		{
			Name:     "Basic Operations",
			Capacity: 2,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"set", "2", "two", nil, false, 0},
				{"get", "1", nil, "one", true, 2},    // Access 1 multiple times
				{"get", "2", nil, "two", true, 1},    // Access 2 once
				{"set", "3", "three", nil, false, 0}, // Should evict 2 (lowest frequency)
				{"get", "2", nil, nil, false, 0},     // 2 should be evicted
				{"get", "1", nil, "one", true, 1},
				{"get", "3", nil, "three", true, 1},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Tie breaking - LRU within same frequency",
			Capacity: 2,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"set", "2", "two", nil, false, 0},
				{"get", "1", nil, "one", true, 1},    // Access 1 once
				{"get", "2", nil, "two", true, 1},    // Access 2 once
				{"set", "3", "three", nil, false, 0}, // Should evict 1 (older than 2)
				{"get", "1", nil, nil, false, 0},     // 1 should be evicted
				{"get", "2", nil, "two", true, 1},
				{"get", "3", nil, "three", true, 1},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Update existing key increases frequency",
			Capacity: 2,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"set", "2", "two", nil, false, 0},
				{"set", "1", "one_updated", nil, false, 0}, // Update 1 (frequency increases)
				{"set", "3", "three", nil, false, 0},       // Should evict 2 (lower frequency)
				{"get", "2", nil, nil, false, 0},           // 2 should be evicted
				{"get", "1", nil, "one_updated", true, 1},
				{"get", "3", nil, "three", true, 1},
			},
			ExpectedLen: 2,
		},
		{
			Name:     "Remove operation",
			Capacity: 2,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"set", "2", "two", nil, false, 0},
				{"remove", "1", nil, nil, false, 0},
				{"get", "1", nil, nil, false, 0}, // 1 should be gone
				{"get", "2", nil, "two", true, 1},
			},
			ExpectedLen: 1,
		},
		{
			Name:     "Purge operation",
			Capacity: 3,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"set", "2", "two", nil, false, 0},
				{"set", "3", "three", nil, false, 0},
				{"purge", "", nil, nil, false, 0},
				{"get", "1", nil, nil, false, 0}, // All should be gone
				{"get", "2", nil, nil, false, 0},
				{"get", "3", nil, nil, false, 0},
			},
			ExpectedLen: 0,
		},
		{
			Name:     "Single item capacity",
			Capacity: 1,
			Operations: []LFUOperation{
				{"set", "1", "one", nil, false, 0},
				{"get", "1", nil, "one", true, 3},  // Access multiple times
				{"set", "2", "two", nil, false, 0}, // Should evict 1
				{"get", "1", nil, nil, false, 0},   // 1 should be evicted
				{"get", "2", nil, "two", true, 1},
			},
			ExpectedLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewLFUCache(tc.Capacity)

			for _, op := range tc.Operations {
				switch op.Type {
				case "set":
					c.Set(op.Key, op.Value)
				case "get":
					for i := 0; i < op.AccessCount; i++ {
						val, ok := c.Get(op.Key)
						if i == op.AccessCount-1 { // Only check on last access
							assert.Equal(t, op.ExpectedOK, ok)
							assert.Equal(t, op.Expected, val)
						}
					}
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
