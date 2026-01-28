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

func TestArray_SetGet(t *testing.T) {
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
			key:           "test_array",
			values:        []string{"test_value"},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set and get multiple values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{"one", "two", "three", "four"},
			ttl:           time.Minute * 30,
			expectedError: false,
		},
		{
			name: "success set with no TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{"no_ttl"},
			ttl:           0,
			expectedError: false,
		},
		{
			name: "set empty values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "set with very long TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{"test_value"},
			ttl:           time.Hour * 24 * 30, // 30 days
			expectedError: false,
		},
		{
			name: "set with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{"value with spaces", "special\nchars\t\r\n"},
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "set with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_array",
			values:        []string{"unicode: 🚀 🎉 📊"},
			ttl:           time.Minute,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			a := NewArray[string](client)
			testKey := uuid.New().String()

			// act
			err := a.Set(t.Context(), testKey, tt.values, tt.ttl)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				got, err := a.Get(t.Context(), testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.values, got)
			}
		})
	}
}

func TestArray_Integers(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	a := NewArray[int64](client)
	key := uuid.New().String()
	val := []int64{10, 20, 30}

	err := a.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	got, err := a.Get(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, val, got)
}

func TestArray_Delete(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	a := NewArray[string](client)
	key := uuid.New().String()
	val := []string{"to_delete"}

	err := a.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	err = a.Delete(t.Context(), key)
	require.NoError(t, err)

	got, err := a.Get(t.Context(), key)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestArray_Pop(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	a := NewArray[string](client)
	key := uuid.New().String()
	val := []string{"pop_me"}

	err := a.Set(t.Context(), key, val, time.Minute)
	require.NoError(t, err)

	got, err := a.Pop(t.Context(), key)
	require.NoError(t, err)
	assert.Equal(t, val, got)

	// Should be deleted (empty)
	res, err := a.Get(t.Context(), key)
	require.NoError(t, err)
	assert.Empty(t, res)
}
