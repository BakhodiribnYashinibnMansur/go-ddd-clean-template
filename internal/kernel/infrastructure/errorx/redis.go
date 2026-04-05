package errorx

import (
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

// HandleRedisError handles Redis errors and converts them to AppError
// This centralizes all Redis error handling logic
func HandleRedisError(err error, key string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	if errors.Is(err, redis.Nil) {
		return handleRedisNil(key, extraFields)
	}

	errMsg := err.Error()

	// Try specific error types
	if appErr := tryHandleSpecificRedisErrors(errMsg, key, extraFields); appErr != nil {
		return appErr
	}

	// Default: Generic Redis error
	appErr := AutoSource(WrapRepoError(err, ErrRepoDatabase, "redis operation failed"))
	_ = appErr.WithField("key", key)
	_ = appErr.WithDetails(errMsg)

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisNil(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(NewRepoError(ErrRepoNotFound, "key not found in cache"))
	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithDetails("The requested key does not exist in Redis")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func tryHandleSpecificRedisErrors(errMsg, key string, extraFields map[string]any) *AppError {
	if isRedisTimeoutError(errMsg) {
		return handleRedisTimeoutError(key, extraFields)
	}
	if isRedisConnectionError(errMsg) {
		return handleRedisConnectionError(key, extraFields)
	}
	if isRedisAuthError(errMsg) {
		return handleRedisAuthError(key, extraFields)
	}
	if isRedisTypeError(errMsg) {
		return handleRedisTypeError(key, extraFields)
	}
	if isRedisMemoryError(errMsg) {
		return handleRedisMemoryError(key, extraFields)
	}
	if isRedisReadOnlyError(errMsg) {
		return handleRedisReadOnlyError(key, extraFields)
	}
	if isRedisClusterError(errMsg) {
		return handleRedisClusterError(key, extraFields)
	}
	if isRedisNoScriptError(errMsg) {
		return handleRedisNoScriptError(key, extraFields)
	}
	if isRedisNoAuthError(errMsg) {
		return handleRedisNoAuthError(key, extraFields)
	}
	return nil
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

func handleRedisConnectionError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoConnection,
			"redis connection error"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "connection")
	_ = appErr.WithDetails("Failed to connect to Redis server")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisTimeoutError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoTimeout,
			"redis operation timeout"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "timeout")
	_ = appErr.WithDetails("Redis operation exceeded timeout limit")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisAuthError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis authentication failed"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "auth")
	_ = appErr.WithDetails("Invalid Redis password or authentication failed")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisTypeError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis wrong type error"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "wrongtype")
	_ = appErr.WithDetails("Operation against a key holding the wrong kind of value")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisMemoryError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis out of memory"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "oom")
	_ = appErr.WithDetails("Redis server is out of memory")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisReadOnlyError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis is read-only"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "readonly")
	_ = appErr.WithDetails("Redis server is in read-only mode (replica)")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisClusterError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis cluster error"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "cluster")
	_ = appErr.WithDetails("Redis cluster is down or key moved to another node")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisNoScriptError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis script not found"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "noscript")
	_ = appErr.WithDetails("Lua script not found in Redis cache")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	return appErr
}

func handleRedisNoAuthError(key string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"redis authentication required"))

	if key != "" {
		_ = appErr.WithField("key", key)
	}
	_ = appErr.WithField("error_type", "noauth")
	_ = appErr.WithDetails("Redis server requires authentication")

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
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
