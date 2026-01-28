package store

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_SetGet(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		key           string
		values        []string
		ttl           time.Duration
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success set and get single value",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"single_value"},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set and get multiple values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"value1", "value2", "value3", "value4"},
			ttl:           time.Minute * 30,
			expectedError: false,
		},
		{
			name: "success set with no TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"no_ttl_value"},
			ttl:           0,
			expectedError: false,
		},
		{
			name: "success set empty values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"value with spaces", "special\nchars\t\r\n"},
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "success set with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"unicode: 🚀 🎉 📊"},
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "set with very long TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_list",
			values:        []string{"test_value"},
			ttl:           time.Hour * 24 * 30, // 30 days
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			l := NewList[string](client)
			testKey := uuid.New().String()

			// act
			err := l.Set(t.Context(), testKey, tt.values, tt.ttl)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				got, err := l.Get(t.Context(), testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.values, got)
			}
		})
	}
}

func TestList_SetGetFull(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	l := NewList[string](client)
	key := uuid.New().String()
	val := []string{"a", "b"}

	err := l.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	len, err := l.GetFull(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), len)
}

func TestList_Pop(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	l := NewList[string](client)
	key := uuid.New().String()
	// Set uses RPush to preserve order. "1", "2", "3" -> List: "1", "2", "3"
	val := []string{"1", "2", "3"}

	err := l.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	// Pop with limit 2, offset 0. Should get "1", "2"
	got, err := l.Pop(t.Context(), key, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, got)

	// Original list should be deleted because Pop implementation deletes the key after retrieving range
	// wait, Pop implementation:
	// get := pipe.LRange(ctx, key, offset, offset+limit-1)
	// pipe.Del(ctx, key)
	// So it deletes the WHOLE key, not just popped elements?
	// Yes: pipe.Del(ctx, key)

	exists, err := client.Exists(t.Context(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestList_Len(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	l := NewList[string](client)
	key := uuid.New().String()
	val := []string{"x"}

	err := l.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	n, err := l.Len(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}
