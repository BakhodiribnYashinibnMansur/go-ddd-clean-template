package errors

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

// HandleRedisError handles Redis errors and converts them to AppError
// This centralizes all Redis error handling logic
func HandleRedisError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	// ============================================================================
	// Special Case: redis.Nil (Key not found)
	// ============================================================================
	if errors.Is(err, redis.Nil) {
		appErr := AutoSource(
			NewRepoError(ctx, ErrRepoNotFound,
				"key not found in cache"))

		if key != "" {
			appErr.WithField("key", key)
		}
		appErr.WithDetails("The requested key does not exist in Redis")

		for k, value := range extraFields {
			appErr.WithField(k, value)
		}
		return appErr
	}

	// Get error message for pattern matching
	errMsg := err.Error()

	// ============================================================================
	// Connection Errors
	// ============================================================================
	if isRedisConnectionError(errMsg) {
		return handleRedisConnectionError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Timeout Errors
	// ============================================================================
	if isRedisTimeoutError(errMsg) {
		return handleRedisTimeoutError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Authentication Errors
	// ============================================================================
	if isRedisAuthError(errMsg) {
		return handleRedisAuthError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Type Errors (WRONGTYPE)
	// ============================================================================
	if isRedisTypeError(errMsg) {
		return handleRedisTypeError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Memory Errors (OOM)
	// ============================================================================
	if isRedisMemoryError(errMsg) {
		return handleRedisMemoryError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Read-Only Errors (READONLY)
	// ============================================================================
	if isRedisReadOnlyError(errMsg) {
		return handleRedisReadOnlyError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Cluster Errors (CLUSTERDOWN, MOVED, ASK)
	// ============================================================================
	if isRedisClusterError(errMsg) {
		return handleRedisClusterError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// NOSCRIPT Error (Lua script not found)
	// ============================================================================
	if isRedisNoScriptError(errMsg) {
		return handleRedisNoScriptError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// NOAUTH Error (Authentication required)
	// ============================================================================
	if isRedisNoAuthError(errMsg) {
		return handleRedisNoAuthError(ctx, err, key, extraFields)
	}

	// ============================================================================
	// Default: Generic Redis error
	// ============================================================================
	appErr := AutoSource(
		WrapRepoError(ctx, err, ErrRepoDatabase,
			"redis operation failed"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithDetails(errMsg)

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

// ============================================================================
// Error Detection Functions
// ============================================================================

func isRedisConnectionError(msg string) bool {
	return strings.Contains(msg, "connection") ||
		strings.Contains(msg, "dial") ||
		strings.Contains(msg, "connect") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset")
}

func isRedisTimeoutError(msg string) bool {
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "i/o timeout")
}

func isRedisAuthError(msg string) bool {
	return strings.Contains(msg, "WRONGPASS") ||
		strings.Contains(msg, "invalid password") ||
		strings.Contains(msg, "authentication failed")
}

func isRedisTypeError(msg string) bool {
	return strings.Contains(msg, "WRONGTYPE")
}

func isRedisMemoryError(msg string) bool {
	return strings.Contains(msg, "OOM") ||
		strings.Contains(msg, "out of memory") ||
		strings.Contains(msg, "maxmemory")
}

func isRedisReadOnlyError(msg string) bool {
	return strings.Contains(msg, "READONLY") ||
		strings.Contains(msg, "read only")
}

func isRedisClusterError(msg string) bool {
	return strings.Contains(msg, "CLUSTERDOWN") ||
		strings.Contains(msg, "MOVED") ||
		strings.Contains(msg, "ASK")
}

func isRedisNoScriptError(msg string) bool {
	return strings.Contains(msg, "NOSCRIPT")
}

func isRedisNoAuthError(msg string) bool {
	return strings.Contains(msg, "NOAUTH")
}

// ============================================================================
// Error Handler Functions
// ============================================================================

func handleRedisConnectionError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoConnection,
			"redis connection error"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "connection").
		WithDetails("Failed to connect to Redis server")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisTimeoutError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoTimeout,
			"redis operation timeout"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "timeout").
		WithDetails("Redis operation exceeded timeout limit")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisAuthError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis authentication failed"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "auth").
		WithDetails("Invalid Redis password or authentication failed")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisTypeError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis wrong type error"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "wrongtype").
		WithDetails("Operation against a key holding the wrong kind of value")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisMemoryError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis out of memory"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "oom").
		WithDetails("Redis server is out of memory")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisReadOnlyError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis is read-only"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "readonly").
		WithDetails("Redis server is in read-only mode (replica)")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisClusterError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis cluster error"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "cluster").
		WithDetails("Redis cluster is down or key moved to another node")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisNoScriptError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis script not found"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "noscript").
		WithDetails("Lua script not found in Redis cache")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisNoAuthError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ctx, ErrRepoDatabase,
			"redis authentication required"))

	if key != "" {
		appErr.WithField("key", key)
	}
	appErr.WithField("error_type", "noauth").
		WithDetails("Redis server requires authentication")

	for k, value := range extraFields {
		appErr.WithField(k, value)
	}
	return appErr
}

// ============================================================================
// Common Redis Error Messages Reference
// ============================================================================
// https://redis.io/docs/reference/cluster-spec/
// https://redis.io/docs/manual/client-side-caching/
//
// Connection Errors:
//   - "dial tcp: connection refused"
//   - "EOF"
//   - "broken pipe"
//   - "connection reset by peer"
//
// Timeout Errors:
//   - "i/o timeout"
//   - "deadline exceeded"
//   - "context deadline exceeded"
//
// Authentication Errors:
//   - "WRONGPASS invalid username-password pair"
//   - "NOAUTH Authentication required"
//
// Type Errors:
//   - "WRONGTYPE Operation against a key holding the wrong kind of value"
//
// Memory Errors:
//   - "OOM command not allowed when used memory > 'maxmemory'"
//
// Read-Only Errors:
//   - "READONLY You can't write against a read only replica"
//
// Cluster Errors:
//   - "CLUSTERDOWN The cluster is down"
//   - "MOVED <slot> <ip>:<port>" (key moved to another node)
//   - "ASK <slot> <ip>:<port>" (temporary redirect)
//
// Script Errors:
//   - "NOSCRIPT No matching script. Please use EVAL"
//
// Key Not Found:
//   - redis.Nil (special case, not an error message)
//
// Other Common Errors:
//   - "ERR unknown command"
//   - "ERR wrong number of arguments"
//   - "ERR syntax error"
//   - "LOADING Redis is loading the dataset in memory"
//   - "BUSY Redis is busy running a script"
//   - "MISCONF Redis is configured to save RDB snapshots, but it is currently not able to persist on disk"
