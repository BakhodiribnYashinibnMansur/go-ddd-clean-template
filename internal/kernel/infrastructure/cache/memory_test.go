package cache_test

import (
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/cache"
	"github.com/stretchr/testify/assert"
)

type MemoryCacheTestCase struct {
	Name        string
	Key         string
	Value       any
	Duration    time.Duration
	WaitTime    time.Duration
	ExpectedGet any
	ExpectedOK  bool
	SetupFunc   func(*cache.MemoryCache)
}

func TestMemoryCache_TableDriven(t *testing.T) {
	testCases := []MemoryCacheTestCase{
		{
			Name:        "Set and Get - success",
			Key:         "key1",
			Value:       "value1",
			Duration:    0,
			WaitTime:    0,
			ExpectedGet: "value1",
			ExpectedOK:  true,
		},
		{
			Name:        "Set and Get - with expiration",
			Key:         "key2",
			Value:       "value2",
			Duration:    10 * time.Millisecond,
			WaitTime:    0,
			ExpectedGet: "value2",
			ExpectedOK:  true,
		},
		{
			Name:        "Set and Get - expired",
			Key:         "key3",
			Value:       "value3",
			Duration:    1 * time.Millisecond,
			WaitTime:    5 * time.Millisecond,
			ExpectedGet: nil,
			ExpectedOK:  false,
		},
		{
			Name:        "Get - missing key",
			Key:         "missing",
			Value:       nil,
			Duration:    0,
			WaitTime:    0,
			ExpectedGet: nil,
			ExpectedOK:  false,
		},
		{
			Name:        "Set and Get - update existing",
			Key:         "key4",
			Value:       "initial",
			Duration:    0,
			WaitTime:    0,
			ExpectedGet: "newValue4",
			ExpectedOK:  true,
			SetupFunc: func(c *cache.MemoryCache) {
				c.Set("key4", "newValue4", 0)
			},
		},
		{
			Name:        "Set and Get - complex type",
			Key:         "complex",
			Value:       struct{ Name string }{Name: "test"},
			Duration:    0,
			WaitTime:    0,
			ExpectedGet: struct{ Name string }{Name: "test"},
			ExpectedOK:  true,
		},
		{
			Name:        "Set and Get - zero duration (no expiration)",
			Key:         "no_exp",
			Value:       "no_expire",
			Duration:    0,
			WaitTime:    0,
			ExpectedGet: "no_expire",
			ExpectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewMemoryCache()

			if tc.Value != nil {
				c.Set(tc.Key, tc.Value, tc.Duration)
			}

			if tc.SetupFunc != nil {
				tc.SetupFunc(c)
			}

			if tc.WaitTime > 0 {
				time.Sleep(tc.WaitTime)
			}

			val, ok := c.Get(tc.Key)

			assert.Equal(t, tc.ExpectedOK, ok)
			assert.Equal(t, tc.ExpectedGet, val)
		})
	}
}

type MemoryCacheDeleteTestCase struct {
	Name        string
	SetupFunc   func(*cache.MemoryCache)
	DeleteKey   string
	ExpectedGet any
	ExpectedOK  bool
}

func TestMemoryCache_Delete_TableDriven(t *testing.T) {
	testCases := []MemoryCacheDeleteTestCase{
		{
			Name: "Delete - existing key",
			SetupFunc: func(c *cache.MemoryCache) {
				c.Set("key1", "value1", 0)
			},
			DeleteKey:   "key1",
			ExpectedGet: nil,
			ExpectedOK:  false,
		},
		{
			Name: "Delete - non-existing key",
			SetupFunc: func(c *cache.MemoryCache) {
				c.Set("key1", "value1", 0)
			},
			DeleteKey:   "missing",
			ExpectedGet: "value1",
			ExpectedOK:  true,
		},
		{
			Name:        "Delete - empty cache",
			SetupFunc:   func(c *cache.MemoryCache) {},
			DeleteKey:   "any_key",
			ExpectedGet: nil,
			ExpectedOK:  false,
		},
		{
			Name: "Delete - multiple keys then delete one",
			SetupFunc: func(c *cache.MemoryCache) {
				c.Set("key1", "value1", 0)
				c.Set("key2", "value2", 0)
				c.Set("key3", "value3", 0)
			},
			DeleteKey:   "key2",
			ExpectedGet: "value1",
			ExpectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			c := cache.NewMemoryCache()

			tc.SetupFunc(c)

			c.Delete(tc.DeleteKey)

			// Check that the deleted key is gone
			_, ok := c.Get(tc.DeleteKey)
			assert.False(t, ok)

			// If we expected another key to still exist, check it
			if tc.ExpectedGet != nil {
				val, ok := c.Get("key1")
				assert.Equal(t, tc.ExpectedOK, ok)
				assert.Equal(t, tc.ExpectedGet, val)
			}
		})
	}
}
