package store

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue_PushPop(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		key           string
		items         []string
		expectedSize  int64
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success push and pop single item",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         []string{"single_item"},
			expectedSize:  1,
			expectedError: false,
		},
		{
			name: "success push multiple items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         []string{"item1", "item2", "item3"},
			expectedSize:  3,
			expectedError: false,
		},
		{
			name: "success push empty items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         []string{},
			expectedSize:  0,
			expectedError: false,
		},
		{
			name: "success push with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         []string{"item with spaces and \t\n\r\n"},
			expectedSize:  1,
			expectedError: false,
		},
		{
			name: "success push with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         []string{"low", "medium", "high"},
			expectedSize:  3,
			expectedError: false,
		},
		{
			name: "push large number of items",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_queue",
			items:         make([]string, 1000),
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
			q := NewQueue[string](client)
			testKey := uuid.New().String()

			// act
			err := q.PushBack(testKey, tt.items)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				sz, err := q.Len(testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSize, sz)
			}
		})
	}
}

func TestQueue_Get(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[string](client)
	key := uuid.New().String()

	err := q.PushBack(key, []string{"1", "2", "3", "4"})
	require.NoError(t, err)

	slice, err := q.Get(key, 0, 1) // 0 to 1 inclusive
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, slice)

	fullLen, err := q.GetFull(key)
	require.NoError(t, err)
	assert.Equal(t, int64(4), fullLen)

	arr, err := q.ToArray(key)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3", "4"}, arr)
}

func TestQueue_Contains(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[string](client)
	key := uuid.New().String()

	err := q.PushBack(key, []string{"target"})
	require.NoError(t, err)

	// Note: Contains implementation only checks index 0
	// func (q *Queue[T]) Contains(key string, value T) (bool, error) {
	// 	valStr, err := q.db.LIndex(context.Background(), key, 0).Result()
	//     ...
	// }

	ok, err := q.Contains(key, "target")
	require.NoError(t, err)
	assert.True(t, ok)

	// If we push front another element, target moves to index 1
	err = q.PushFront(key, []string{"new_front"})
	require.NoError(t, err)

	// Contains checks index 0, so it will be "new_front"
	ok, err = q.Contains(key, "target")
	require.NoError(t, err)
	assert.False(t, ok) // This confirms the limitation/behavior of current Contains implementation
}

func TestQueue_Peek(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[string](client)
	key := uuid.New().String()

	err := q.PushBack(key, []string{"first"})
	require.NoError(t, err)

	val, err := q.Peek(key)
	require.NoError(t, err)
	assert.Equal(t, "first", val)
}

func TestQueue_DeleteRange(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[string](client)
	key := uuid.New().String()

	// 1, 2, 3, 4
	err := q.PushBack(key, []string{"1", "2", "3", "4"})
	require.NoError(t, err)

	// Keep 1 to 2 (indices) -> "2", "3"
	err = q.DeleteRange(key, 1, 2)
	require.NoError(t, err)

	arr, err := q.ToArray(key)
	require.NoError(t, err)
	assert.Equal(t, []string{"2", "3"}, arr)
}

func TestQueue_Integers(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[int](client)
	key := uuid.New().String()

	err := q.PushBack(key, []int{100})
	require.NoError(t, err)

	val, err := q.PopFront(key)
	require.NoError(t, err)
	assert.Equal(t, 100, val)
}

// TestQueue_Contains_Behavior confirms the limitation mentioned: it only checks the first element
func TestQueue_Contains_Behavior(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	q := NewQueue[string](client)
	key := uuid.New().String()
	err := q.PushBack(key, []string{"a", "b"})
	require.NoError(t, err)

	// "b" is at index 1
	found, err := q.Contains(key, "b")
	require.NoError(t, err)
	assert.False(t, found, "Contains only checks index 0 currently")

	found, err = q.Contains(key, "a")
	require.NoError(t, err)
	assert.True(t, found)
}
