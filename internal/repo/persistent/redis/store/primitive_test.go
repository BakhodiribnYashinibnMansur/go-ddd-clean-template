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

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestPrimitive_SetGet(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		key           string
		value         string
		ttl           time.Duration
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success set and get single value",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "test_value",
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set with no TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "test_value",
			ttl:           0,
			expectedError: false,
		},
		{
			name: "success set with very long TTL",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "test_value",
			ttl:           time.Hour * 24 * 30, // 30 days
			expectedError: false,
		},
		{
			name: "success set empty value",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "",
			ttl:           time.Hour,
			expectedError: false,
		},
		{
			name: "success set with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "value with spaces and \t\n\r",
			ttl:           time.Minute,
			expectedError: false,
		},
		{
			name: "success set with unicode characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			key:           "test_primitive",
			value:         "unicode: 🚀 🎉 📊",
			ttl:           time.Minute,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			p := NewPrimitive[string](client)
			testKey := uuid.New().String()

			// act
			err := p.Set(testKey, tt.value, tt.ttl)

			// assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				got, err := p.Get(testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			}
		})
	}
}

func TestPrimitive_Integers(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	p := NewPrimitive[int64](client)
	key := uuid.New().String()
	val := int64(12345)

	err := p.Set(key, val, time.Minute)
	require.NoError(t, err)

	got, err := p.Get(key)
	require.NoError(t, err)
	assert.Equal(t, val, got)
}

func TestPrimitive_Delete(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	p := NewPrimitive[string](client)
	key := uuid.New().String()
	val := "to_delete"

	err := p.Set(key, val, time.Minute)
	require.NoError(t, err)

	err = p.Delete(key)
	require.NoError(t, err)

	_, err = p.Get(key)
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestPrimitive_Pop(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	p := NewPrimitive[string](client)
	key := uuid.New().String()
	val := "pop_me"

	err := p.Set(key, val, time.Minute)
	require.NoError(t, err)

	got, err := p.Pop(key)
	require.NoError(t, err)
	assert.Equal(t, val, got)

	// Should be deleted
	exists, err := p.Exists(key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPrimitive_Exists(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	p := NewPrimitive[string](client)
	key := uuid.New().String()

	exists, err := p.Exists(key)
	require.NoError(t, err)
	assert.False(t, exists)

	err = p.Set(key, "val", time.Minute)
	require.NoError(t, err)

	exists, err = p.Exists(key)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestPrimitive_Scan(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	p := NewPrimitive[string](client)
	prefix := uuid.New().String()

	keys := []string{
		prefix + ":1",
		prefix + ":2",
		prefix + ":3",
	}

	for _, k := range keys {
		err := p.Set(k, "val", time.Minute)
		require.NoError(t, err)
	}

	foundKeys, err := p.Scan(prefix + "*")
	require.NoError(t, err)
	assert.Len(t, foundKeys, 3)
	assert.ElementsMatch(t, keys, foundKeys)
}
