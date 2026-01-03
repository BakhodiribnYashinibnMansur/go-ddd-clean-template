package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriorityQueue_PushPop(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		key           string
		items         []GenericZ[string]
		expectedSize  int64
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success push and pop single item",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key: "test_priority_queue",
			items: []GenericZ[string]{
				{Score: 10, Member: "low"},
			},
			expectedSize:  1,
			expectedError: false,
		},
		{
			name: "success push multiple items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key: "test_priority_queue",
			items: []GenericZ[string]{
				{Score: 10, Member: "low"},
				{Score: 20, Member: "medium"},
				{Score: 30, Member: "high"},
			},
			expectedSize:  3,
			expectedError: false,
		},
		{
			name: "success push empty items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_priority_queue",
			items:         []GenericZ[string]{},
			expectedSize:  0,
			expectedError: false,
		},
		{
			name: "success push with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key: "test_priority_queue",
			items: []GenericZ[string]{
				{Score: 10, Member: "item with spaces and \t\n\r\n"},
			},
			expectedSize:  1,
			expectedError: false,
		},
		{
			name: "success push with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key: "test_priority_queue",
			items: []GenericZ[string]{
				{Score: 10, Member: "unicode: 🚀 🎉 📊"},
			},
			expectedSize:  1,
			expectedError: false,
		},
		{
			name: "push large number of items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key: "test_priority_queue",
			items: func() []GenericZ[string] {
				items := make([]GenericZ[string], 1000)
				for i := range items {
					items[i] = GenericZ[string]{Score: float64(i), Member: fmt.Sprintf("item_%d", i)}
				}
				return items
			}(),
			expectedSize:  1000,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			pq := NewPriorityQueue[string](client)
			testKey := uuid.New().String()

			// act
			err := pq.Push(testKey, tt.items)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				sz, err := pq.Size(testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSize, sz)
			}
		})
	}
}

func TestPriorityQueue_Get(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	pq := NewPriorityQueue[string](client)
	key := uuid.New().String()

	items := []GenericZ[string]{
		{Score: 1, Member: "one"},
		{Score: 2, Member: "two"},
		{Score: 3, Member: "three"},
	}
	err := pq.Push(key, items)
	require.NoError(t, err)

	// Get range 0-1
	slice, err := pq.Get(key, 0, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, slice)

	// ToArray (all)
	all, err := pq.ToArray(key)
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, all)
}

func TestPriorityQueue_ChangePriority(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	pq := NewPriorityQueue[string](client)
	key := uuid.New().String()

	items := []GenericZ[string]{
		{Score: 10, Member: "item"},
	}
	err := pq.Push(key, items)
	require.NoError(t, err)

	err = pq.ChangePriority(key, "item", 100)
	require.NoError(t, err)

	// Check if score updated (by checking order if we had multiple, or just by popping)
	// We can check Score using direct redis client commands or infer from behavior

	score, err := client.ZScore(context.Background(), key, "item").Result()
	require.NoError(t, err)
	assert.Equal(t, float64(100), score)
}

func TestPriorityQueue_DeleteRange(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	pq := NewPriorityQueue[string](client)
	key := uuid.New().String()

	items := []GenericZ[string]{
		{Score: 1, Member: "1"},
		{Score: 2, Member: "2"},
		{Score: 3, Member: "3"},
	}
	err := pq.Push(key, items)
	require.NoError(t, err)

	// Delete rank 0 (lowest score: "1")
	err = pq.DeleteRange(key, 0, 0)
	require.NoError(t, err)

	all, err := pq.ToArray(key)
	require.NoError(t, err)
	assert.Equal(t, []string{"2", "3"}, all)
}

func TestPriorityQueue_Clear(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	pq := NewPriorityQueue[string](client)
	key := uuid.New().String()

	pq.Push(key, []GenericZ[string]{{Score: 1, Member: "a"}})

	err := pq.Clear(key)
	require.NoError(t, err)

	empty, err := pq.IsEmpty(key)
	require.NoError(t, err)
	assert.True(t, empty)
}
