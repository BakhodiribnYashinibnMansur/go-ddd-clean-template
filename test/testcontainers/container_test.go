package container

import (
	"testing"
	"time"

	"gct/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type containerTestCase struct {
	name        string
	setupFunc   func() any
	testFunc    func(any) error
	expectError bool
	definition  string
}

func TestContainer_TableDriven(t *testing.T) {
	tests := []containerTestCase{
		{
			name: "Redis miniredis setup",
			setupFunc: func() any {
				mr, err := miniredis.Run()
				require.NoError(t, err)
				t.Cleanup(mr.Close)

				cfg := config.RedisStore{
					Host: mr.Host(),
					Port: mr.Port(),
					DB:   0,
				}

				client := redis.NewClient(&redis.Options{
					Addr: cfg.Addr(),
					DB:   cfg.DB,
				})

				return client
			},
			testFunc: func(container any) error {
				client := container.(*redis.Client)
				ctx := t.Context()
				return client.Ping(ctx).Err()
			},
			expectError: false,
			definition:  "Tests Redis miniredis container setup and connection",
		},
		{
			name: "Redis configuration",
			setupFunc: func() any {
				cfg := config.RedisStore{
					Host: "localhost",
					Port: "6379",
					DB:   1,
				}
				return cfg
			},
			testFunc: func(container any) error {
				cfg := container.(config.RedisStore)
				expectedAddr := "localhost:6379"
				if cfg.Addr() != expectedAddr {
					return assert.AnError
				}
				if cfg.DB != 1 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis configuration address formatting",
		},
		{
			name: "Redis configuration defaults",
			setupFunc: func() any {
				cfg := config.RedisStore{}
				return cfg
			},
			testFunc: func(container any) error {
				cfg := container.(config.RedisStore)
				expectedAddr := "localhost:6379"
				if cfg.Addr() != expectedAddr {
					return assert.AnError
				}
				if cfg.DB != 0 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis configuration default values",
		},
		{
			name: "Minio configuration",
			setupFunc: func() any {
				cfg := config.MinioStore{
					AccessKey: "testkey",
					SecretKey: "testsecret",
					Bucket:    "testbucket",
					UseSSL:    false,
				}
				return cfg
			},
			testFunc: func(container any) error {
				cfg := container.(config.MinioStore)
				if cfg.AccessKey != "testkey" {
					return assert.AnError
				}
				if cfg.SecretKey != "testsecret" {
					return assert.AnError
				}
				if cfg.Bucket != "testbucket" {
					return assert.AnError
				}
				if cfg.UseSSL != false {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Minio configuration values",
		},
		{
			name: "Constants validation",
			setupFunc: func() any {
				constants := struct {
					MinioImage string
					RedisImage string
				}{
					MinioImage: MinioImage,
					RedisImage: RedisImage,
				}
				return constants
			},
			testFunc: func(container any) error {
				constants := container.(struct {
					MinioImage string
					RedisImage string
				})
				if constants.MinioImage == "" {
					return assert.AnError
				}
				if constants.RedisImage == "" {
					return assert.AnError
				}
				if constants.MinioImage != "minio/minio:latest" {
					return assert.AnError
				}
				if constants.RedisImage != "redis:7-alpine" {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests container image constants",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			container := tc.setupFunc()
			err := tc.testFunc(container)

			if tc.expectError {
				assert.Error(t, err, "Expected test to fail")
			} else {
				assert.NoError(t, err, "Expected test to pass")
			}
		})
	}
}

// TestRedisOperations tests Redis operations using miniredis
func TestRedisOperations_TableDriven(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { client.Close() })

	type redisOperationTestCase struct {
		name        string
		setupKey    string
		setupValue  string
		operation   func(*redis.Client) error
		expectError bool
		definition  string
	}

	tests := []redisOperationTestCase{
		{
			name:       "SET and GET operation",
			setupKey:   "test:key",
			setupValue: "test-value",
			operation: func(rdb *redis.Client) error {
				ctx := t.Context()
				err := rdb.Set(ctx, "test:key", "test-value", time.Hour).Err()
				if err != nil {
					return err
				}
				val, err := rdb.Get(ctx, "test:key").Result()
				if err != nil {
					return err
				}
				if val != "test-value" {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests basic Redis SET and GET operations",
		},
		{
			name:       "DELETE operation",
			setupKey:   "delete:me",
			setupValue: "to-be-deleted",
			operation: func(rdb *redis.Client) error {
				ctx := t.Context()
				// First set the value
				err := rdb.Set(ctx, "delete:me", "to-be-deleted", time.Hour).Err()
				if err != nil {
					return err
				}
				// Then delete it
				err = rdb.Del(ctx, "delete:me").Err()
				if err != nil {
					return err
				}
				// Verify it's gone
				_, err = rdb.Get(ctx, "delete:me").Result()
				if err == nil {
					return assert.AnError // Should have failed
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis DELETE operation",
		},
		{
			name:       "EXISTS operation",
			setupKey:   "exists:key",
			setupValue: "exists-value",
			operation: func(rdb *redis.Client) error {
				ctx := t.Context()
				// Set the value
				err := rdb.Set(ctx, "exists:key", "exists-value", time.Hour).Err()
				if err != nil {
					return err
				}
				// Check if it exists
				exists, err := rdb.Exists(ctx, "exists:key").Result()
				if err != nil {
					return err
				}
				if exists != 1 {
					return assert.AnError
				}
				// Check non-existent key
				exists, err = rdb.Exists(ctx, "nonexistent:key").Result()
				if err != nil {
					return err
				}
				if exists != 0 {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
			definition:  "Tests Redis EXISTS operation",
		},
		{
			name:       "TTL operation",
			setupKey:   "ttl:key",
			setupValue: "ttl-value",
			operation: func(rdb *redis.Client) error {
				ctx := t.Context()
				// Set with TTL
				err := rdb.Set(ctx, "ttl:key", "ttl-value", 2*time.Second).Err()
				if err != nil {
					return err
				}
				// Check TTL
				ttl, err := rdb.TTL(ctx, "ttl:key").Result()
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation(client)

			if tc.expectError {
				assert.Error(t, err, "Expected operation to fail")
			} else {
				assert.NoError(t, err, "Expected operation to pass")
			}
		})
	}
}

// TestContainerIntegration tests container integration scenarios
func TestContainerIntegration_TableDriven(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	type integrationTestCase struct {
		name        string
		container   string
		testFunc    func() error
		expectError bool
		skip        bool
		skipReason  string
		definition  string
	}

	tests := []integrationTestCase{
		{
			name:      "Redis container startup",
			container: "redis",
			testFunc: func() error {
				// This would test actual Redis container startup
				// but we skip it for now to avoid testcontainers dependency
				return nil
			},
			expectError: false,
			skip:        true,
			skipReason:  "Requires testcontainers setup",
			definition:  "Tests Redis container startup and connectivity",
		},
		{
			name:      "Minio container startup",
			container: "minio",
			testFunc: func() error {
				// This would test actual Minio container startup
				// but we skip it for now to avoid testcontainers dependency
				return nil
			},
			expectError: false,
			skip:        true,
			skipReason:  "Requires testcontainers setup",
			definition:  "Tests Minio container startup and connectivity",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip(tc.skipReason)
			}

			err := tc.testFunc()

			if tc.expectError {
				assert.Error(t, err, "Expected integration test to fail")
			} else {
				assert.NoError(t, err, "Expected integration test to pass")
			}
		})
	}
}
