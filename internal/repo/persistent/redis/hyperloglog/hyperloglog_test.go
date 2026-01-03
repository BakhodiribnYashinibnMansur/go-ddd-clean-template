package hyperloglog

import (
	"context"
	"testing"

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

func TestHyperLogLog_PFAddCount(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		elements      []string
		expectedCount int64
		expectError   bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success add multiple elements",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      []string{"a", "b", "c"},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "success add single element",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      []string{"single"},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "success add empty elements",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      []string{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "success add duplicate elements",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      []string{"a", "a", "b"},
			expectedCount: 2, // HyperLogLog counts unique elements
			expectError:   false,
		},
		{
			name: "add elements with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      []string{"test-item", "user:123", "data@#&*"},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "add large number of elements",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			elements:      make([]string, 1000), // Large dataset
			expectedCount: 1000,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			h := New(client)
			testKey := uuid.New().String()
			testCtx := context.Background()

			// act
			elements := make([]any, len(tt.elements))
			for i, el := range tt.elements {
				elements[i] = el
			}
			_, err := h.PFAdd(testCtx, testKey, elements...)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				// Get the total count using PFCount
				count, err := h.PFCount(testCtx, testKey)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestHyperLogLog_PFMerge(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		sourceKeys    []string
		destKey       string
		expectedCount int64
		expectError   bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success merge two hyperloglogs",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			sourceKeys:    []string{"hll1", "hll2"},
			destKey:       "hll_dest",
			expectedCount: 4,
			expectError:   false,
		},
		{
			name: "merge empty source",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			sourceKeys:    []string{},
			destKey:       "hll_dest",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "merge single source",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			sourceKeys:    []string{"hll1"},
			destKey:       "hll_dest",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "merge non-existent source",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			sourceKeys:    []string{"non_existent_hll"},
			destKey:       "hll_dest",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "merge with overlapping elements",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			sourceKeys:    []string{"hll1", "hll2"},
			destKey:       "hll_dest",
			expectedCount: 3, // Should count unique elements
			expectError:   false,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			h := New(client)
			testCtx := context.Background()

			// setup source hyperloglogs
			if len(tt.sourceKeys) > 0 {
				h.PFAdd(testCtx, tt.sourceKeys[0], []any{"a", "b"}...)
			}
			if len(tt.sourceKeys) > 1 {
				h.PFAdd(testCtx, tt.sourceKeys[1], []any{"c", "d"}...)
			}

			// act
			err := h.PFMerge(testCtx, tt.destKey, tt.sourceKeys...)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				count, err := h.PFCount(testCtx, tt.destKey)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestHyperLogLog_DeleteExists(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	h := New(client)
	key := uuid.New().String()
	ctx := context.Background()

	exists, err := h.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	h.PFAdd(ctx, key, []any{"item"}...)

	exists, err = h.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	err = h.Delete(ctx, key)
	require.NoError(t, err)

	exists, err = h.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}
