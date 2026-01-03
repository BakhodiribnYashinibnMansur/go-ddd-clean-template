package container

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gct/config"
)

// TestRedisConfig_TableDriven tests Redis configuration scenarios
func TestRedisConfig_TableDriven(t *testing.T) {
	type redisConfigTestCase struct {
		name         string
		config       config.RedisStore
		expectedAddr string
		expectedDB   int
		definition   string
	}

	tests := []redisConfigTestCase{
		{
			name: "Default configuration",
			config: config.RedisStore{
				Host: "",
				Port: "",
				DB:   0,
			},
			expectedAddr: "localhost:6379",
			expectedDB:   0,
			definition:   "Tests Redis configuration with default values",
		},
		{
			name: "Custom host and port",
			config: config.RedisStore{
				Host: "redis.example.com",
				Port: "6380",
				DB:   1,
			},
			expectedAddr: "redis.example.com:6380",
			expectedDB:   1,
			definition:   "Tests Redis configuration with custom host and port",
		},
		{
			name: "Only custom host",
			config: config.RedisStore{
				Host: "custom-redis",
				Port: "",
				DB:   2,
			},
			expectedAddr: "custom-redis:6379",
			expectedDB:   2,
			definition:   "Tests Redis configuration with custom host only",
		},
		{
			name: "Only custom port",
			config: config.RedisStore{
				Host: "",
				Port: "6381",
				DB:   3,
			},
			expectedAddr: "localhost:6381",
			expectedDB:   3,
			definition:   "Tests Redis configuration with custom port only",
		},
		{
			name: "With password",
			config: config.RedisStore{
				Host:     "secure-redis",
				Port:     "6379",
				Password: "secret123",
				DB:       0,
			},
			expectedAddr: "secure-redis:6379",
			expectedDB:   0,
			definition:   "Tests Redis configuration with password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualAddr := tc.config.Addr()
			actualDB := tc.config.DB

			assert.Equal(t, tc.expectedAddr, actualAddr, "Address should match expected")
			assert.Equal(t, tc.expectedDB, actualDB, "DB should match expected")
		})
	}
}

// TestRedisOperationsDetailed_TableDriven tests Redis operations using miniredis
func TestRedisOperationsDetailed_TableDriven(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { client.Close() })

	type redisOperationTestCase struct {
		name        string
		key         string
		value       string
		operation   func(*redis.Client, string, string) error
		expectError bool
		definition  string
	}

	tests := []redisOperationTestCase{
		{
			name:  "SET operation",
			key:   "test:set",
			value: "set-value",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				return rdb.Set(ctx, key, value, time.Hour).Err()
			},
			expectError: false,
			definition:  "Tests Redis SET operation",
		},
		{
			name:  "GET operation after SET",
			key:   "test:get",
			value: "get-value",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				// First set the value
				err := rdb.Set(ctx, key, value, time.Hour).Err()
				if err != nil {
					return err
				}
				// Then get it
				actualValue, err := rdb.Get(ctx, key).Result()
				if err != nil {
					return err
				}
				if actualValue != value {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis GET operation after SET",
		},
		{
			name:  "DELETE operation",
			key:   "test:delete",
			value: "delete-value",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				// First set the value
				err := rdb.Set(ctx, key, value, time.Hour).Err()
				if err != nil {
					return err
				}
				// Then delete it
				return rdb.Del(ctx, key).Err()
			},
			expectError: false,
			definition:  "Tests Redis DELETE operation",
		},
		{
			name:  "EXISTS operation",
			key:   "test:exists",
			value: "exists-value",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				// First set the value
				err := rdb.Set(ctx, key, value, time.Hour).Err()
				if err != nil {
					return err
				}
				// Check if it exists
				count, err := rdb.Exists(ctx, key).Result()
				if err != nil {
					return err
				}
				if count != 1 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis EXISTS operation",
		},
		{
			name:  "TTL operation",
			key:   "test:ttl",
			value: "ttl-value",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				// Set with 2 second TTL
				err := rdb.Set(ctx, key, value, 2*time.Second).Err()
				if err != nil {
					return err
				}
				// Check TTL
				ttl, err := rdb.TTL(ctx, key).Result()
				if err != nil {
					return err
				}
				if ttl <= 0 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis TTL operation",
		},
		{
			name:  "INCR operation",
			key:   "test:incr",
			value: "1",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				// Set initial value
				err := rdb.Set(ctx, key, "1", time.Hour).Err()
				if err != nil {
					return err
				}
				// Increment
				newValue, err := rdb.Incr(ctx, key).Result()
				if err != nil {
					return err
				}
				if newValue != 2 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis INCR operation",
		},
		{
			name:  "GET non-existent key",
			key:   "test:nonexistent",
			value: "",
			operation: func(rdb *redis.Client, key, value string) error {
				ctx := t.Context()
				_, err := rdb.Get(ctx, key).Result()
				if err == nil {
					return assert.AnError // Should have failed
				}
				return nil // Expected error
			},
			expectError: false,
			definition:  "Tests Redis GET operation on non-existent key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation(client, tc.key, tc.value)

			if tc.expectError {
				assert.Error(t, err, "Expected operation to fail")
			} else {
				assert.NoError(t, err, "Expected operation to pass")
			}
		})
	}
}

// TestRedisConnection_TableDriven tests Redis connection scenarios
func TestRedisConnection_TableDriven(t *testing.T) {
	type redisConnectionTestCase struct {
		name        string
		setupFunc   func() (*redis.Client, func())
		expectError bool
		definition  string
	}

	tests := []redisConnectionTestCase{
		{
			name: "Valid miniredis connection",
			setupFunc: func() (*redis.Client, func()) {
				mr, err := miniredis.Run()
				require.NoError(t, err)

				client := redis.NewClient(&redis.Options{
					Addr: mr.Addr(),
				})

				cleanup := func() {
					client.Close()
					mr.Close()
				}

				return client, cleanup
			},
			expectError: false,
			definition:  "Tests valid Redis connection using miniredis",
		},
		{
			name: "Invalid connection",
			setupFunc: func() (*redis.Client, func()) {
				client := redis.NewClient(&redis.Options{
					Addr: "localhost:9999", // Invalid port
				})

				cleanup := func() {
					client.Close()
				}

				return client, cleanup
			},
			expectError: true,
			definition:  "Tests invalid Redis connection scenario",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, cleanup := tc.setupFunc()
			defer cleanup()

			ctx := t.Context()
			err := client.Ping(ctx).Err()

			if tc.expectError {
				assert.Error(t, err, "Expected connection to fail")
			} else {
				assert.NoError(t, err, "Expected connection to succeed")
			}
		})
	}
}
