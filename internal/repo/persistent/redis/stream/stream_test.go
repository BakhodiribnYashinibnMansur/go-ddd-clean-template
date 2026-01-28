package stream

import (
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
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

func TestStream_XAdd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		streamKey   string
		fieldValues map[string]any
		id          string
		expectedID  string
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "success add single field",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			streamKey: "test_stream",
			fieldValues: map[string]any{
				"field1": "value1",
				"field2": "value2",
			},
			id:          "test_id_1",
			expectedID:  "test_id_1",
			expectError: false,
		},
		{
			name: "success add multiple fields",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			streamKey: "test_stream",
			fieldValues: map[string]any{
				"field1": "value1",
				"field2": "value2",
				"field3": "value3",
			},
			id:          "test_id_2",
			expectedID:  "test_id_2",
			expectError: false,
		},
		{
			name: "add with empty field values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			streamKey:   "test_stream",
			fieldValues: map[string]any{},
			id:          "test_id_3",
			expectedID:  "test_id_3",
			expectError: false,
		},
		{
			name: "redis connection error",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			streamKey: "test_stream",
			fieldValues: map[string]any{
				"field1": "value1",
			},
			id:          "test_id_4",
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "add with nil field values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			streamKey: "test_stream",
			fieldValues: map[string]any{
				"field1": nil,
				"field2": nil,
			},
			id:          "test_id_5",
			expectedID:  "test_id_5",
			expectError: false,
		},
		{
			name: "add with simple field values",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			streamKey: "test_stream",
			fieldValues: map[string]any{
				"field1": "simple_value",
				"field2": 123,
			},
			id:          "test_id_7",
			expectedID:  "test_id_7",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			s := New(client)
			testCtx := t.Context()

			// act
			var id string
			var err error
			if tt.id == "" {
				id, err = s.XAdd(testCtx, tt.streamKey, tt.fieldValues)
			} else {
				id, err = s.XAddWithID(testCtx, tt.streamKey, tt.id, tt.fieldValues)
			}

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
