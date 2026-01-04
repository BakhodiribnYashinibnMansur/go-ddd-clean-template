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

func TestHashTable_SetGet(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		key           string
		values        map[string]string
		ttl           time.Duration
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success set and get single field",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "value1"},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set and get multiple fields",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			ttl:           time.Minute * 30,
			expectedError: false,
		},
		{
			name: "success set with no TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "value1"},
			ttl:           0,
			expectedError: false,
		},
		{
			name: "success set empty values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{},
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "value with spaces", "field2": "special\nchars\t\r\n"},
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "success set with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "unicode: 🚀 🎉 📊"},
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "set with very long TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_hash",
			values:        map[string]string{"field1": "value1"},
			ttl:           time.Hour * 24 * 30, // 30 days
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			h := NewHashTable[string](client)
			testKey := uuid.New().String()

			// act
			err := h.Set(testKey, tt.values, tt.ttl)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				got, err := h.Get(testKey, false)
				require.NoError(t, err)
				assert.Equal(t, tt.values, got)
			}
		})
	}
}

func TestHashTable_Pop(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	h := NewHashTable[string](client)
	key := uuid.New().String()
	val := map[string]string{"f": "v"}

	err := h.Set(key, val, time.Minute)
	require.NoError(t, err)

	got, err := h.Pop(key)
	require.NoError(t, err)
	assert.Equal(t, val, got)

	// Should be deleted
	exists, err := client.Exists(t.Context(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestHashTable_Delete(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	h := NewHashTable[string](client)
	key := uuid.New().String()
	val := map[string]string{"f": "v"}

	err := h.Set(key, val, time.Minute)
	require.NoError(t, err)

	err = h.Delete(key)
	require.NoError(t, err)

	exists, err := client.Exists(t.Context(), key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}
