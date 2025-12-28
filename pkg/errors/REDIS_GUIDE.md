# Redis Error Handling Guide

## Overview

`pkg/errors/redis.go` provides centralized error handling for all Redis operations, similar to PostgreSQL error handling.

## Core Function

```go
func HandleRedisError(ctx context.Context, err error, key string, extraFields map[string]any) *AppError
```

## Usage Examples

### Basic Usage

```go
// Repository method example
func (r *CacheRepo) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return "", apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "get",
        })
    }
    return val, nil
}
```

### With Extra Context

```go
func (r *CacheRepo) SetUser(ctx context.Context, userID int64, data string) error {
    key := fmt.Sprintf("user:%d", userID)
    
    err := r.client.Set(ctx, key, data, time.Hour).Err()
    if err != nil {
        return apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "set_user",
            "user_id":   userID,
            "ttl":       "1h",
        })
    }
    return nil
}
```

### Hash Operations

```go
func (r *CacheRepo) GetUserField(ctx context.Context, userID int64, field string) (string, error) {
    key := fmt.Sprintf("user:%d", userID)
    
    val, err := r.client.HGet(ctx, key, field).Result()
    if err != nil {
        return "", apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "hget",
            "user_id":   userID,
            "field":     field,
        })
    }
    return val, nil
}
```

### List Operations

```go
func (r *CacheRepo) PushToQueue(ctx context.Context, queueName string, item string) error {
    err := r.client.RPush(ctx, queueName, item).Err()
    if err != nil {
        return apperrors.HandleRedisError(ctx, err, queueName, map[string]any{
            "operation": "rpush",
            "queue":     queueName,
        })
    }
    return nil
}
```

## Error Types Handled

### 1. **Key Not Found (redis.Nil)** → `ErrRepoNotFound`
```go
val, err := rdb.Get(ctx, "nonexistent").Result()
// Returns: ErrRepoNotFound with message "key not found in cache"
```

**Log Output:**
```json
{
  "error_code": "2001",
  "error_type": "REPO_NOT_FOUND",
  "message": "key not found in cache",
  "key": "user:12345",
  "file": "internal/repo/cache/user.go",
  "function": "CacheRepo.Get"
}
```

### 2. **Connection Errors** → `ErrRepoConnection`
- "connection refused"
- "EOF"
- "broken pipe"
- "dial tcp: timeout"

```go
// Automatic detection and handling
err := rdb.Ping(ctx).Err()
// Returns: ErrRepoConnection
```

### 3. **Timeout Errors** → `ErrRepoTimeout`
- "i/o timeout"
- "deadline exceeded"
- "context deadline exceeded"

```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

err := rdb.Get(ctx, "slow-key").Err()
// Returns: ErrRepoTimeout
```

### 4. **Authentication Errors** → `ErrRepoDatabase`
- "WRONGPASS invalid username-password pair"
- "NOAUTH Authentication required"

### 5. **Type Errors (WRONGTYPE)** → `ErrRepoDatabase`
```go
// Key "mykey" is a string
rdb.Set(ctx, "mykey", "value", 0)

// Try to use it as a list
err := rdb.LPush(ctx, "mykey", "item").Err()
// Returns: "WRONGTYPE Operation against a key holding the wrong kind of value"
```

### 6. **Memory Errors (OOM)** → `ErrRepoDatabase`
- "OOM command not allowed when used memory > 'maxmemory'"

### 7. **Read-Only Errors** → `ErrRepoDatabase`
- "READONLY You can't write against a read only replica"

### 8. **Cluster Errors** → `ErrRepoDatabase`
- "CLUSTERDOWN The cluster is down"
- "MOVED <slot> <ip>:<port>"
- "ASK <slot> <ip>:<port>"

### 9. **Script Errors (NOSCRIPT)** → `ErrRepoDatabase`
```go
sha := "nonexistent-sha"
err := rdb.EvalSha(ctx, sha, []string{}, "arg").Err()
// Returns: "NOSCRIPT No matching script"
```

## Complete Repository Example

```go
package cache

import (
    "context"
    "fmt"
    "time"

    apperrors "github.com/evrone/go-clean-template/pkg/errors"
    "github.com/redis/go-redis/v9"
)

type UserCacheRepo struct {
    client *redis.Client
}

func NewUserCacheRepo(client *redis.Client) *UserCacheRepo {
    return &UserCacheRepo{client: client}
}

// Get retrieves user data from cache
func (r *UserCacheRepo) Get(ctx context.Context, userID int64) (string, error) {
    key := fmt.Sprintf("user:%d", userID)
    
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return "", apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "get_user",
            "user_id":   userID,
        })
    }
    
    return val, nil
}

// Set stores user data in cache
func (r *UserCacheRepo) Set(ctx context.Context, userID int64, data string, ttl time.Duration) error {
    key := fmt.Sprintf("user:%d", userID)
    
    err := r.client.Set(ctx, key, data, ttl).Err()
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "set_user",
            "user_id":   userID,
            "ttl":       ttl.String(),
        })
    }
    
    return nil
}

// Delete removes user data from cache
func (r *UserCacheRepo) Delete(ctx context.Context, userID int64) error {
    key := fmt.Sprintf("user:%d", userID)
    
    err := r.client.Del(ctx, key).Err()
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "delete_user",
            "user_id":   userID,
        })
    }
    
    return nil
}

// Exists checks if user exists in cache
func (r *UserCacheRepo) Exists(ctx context.Context, userID int64) (bool, error) {
    key := fmt.Sprintf("user:%d", userID)
    
    count, err := r.client.Exists(ctx, key).Result()
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return false, apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "exists_user",
            "user_id":   userID,
        })
    }
    
    return count > 0, nil
}
```

## Error Flow

```
Repository (Redis)
    ↓
HandleRedisError()  ← Detects error type automatically
    ↓
Returns AppError with proper code
    ↓
Service Layer
    ↓
MapRepoToServiceError()
    ↓
Controller Layer
    ↓
MapServiceToHandlerError() + LOG with zap.Errorw()
```

## Benefits

### ✅ Automatic Error Detection
No need to manually check error messages - `HandleRedisError` does it automatically!

### ✅ Consistent Error Codes
- `ErrRepoNotFound` (2001) - Key not found
- `ErrRepoConnection` (2005) - Connection errors
- `ErrRepoTimeout` (2004) - Timeout errors
- `ErrRepoDatabase` (2003) - All other Redis errors

### ✅ Rich Context
Every error includes:
- `key` - Redis key that failed
- `operation` - What operation was being performed
- `file` - Source file (automatic via AutoSource)
- `function` - Function name (automatic via AutoSource)
- Any extra fields you provide

### ✅ No Logging in Repository
Just like PostgreSQL handler, Redis handler doesn't log - only returns structured errors.

## Comparison: Before vs After

### ❌ Before (Manual):
```go
func (r *Repo) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            r.logger.Error("key not found", zap.String("key", key))
            return "", fmt.Errorf("key %s not found", key)
        }
        if strings.Contains(err.Error(), "timeout") {
            r.logger.Error("timeout", zap.Error(err))
            return "", fmt.Errorf("redis timeout: %w", err)
        }
        r.logger.Error("redis error", zap.Error(err))
        return "", fmt.Errorf("redis error: %w", err)
    }
    return val, nil
}
```

### ✅ After (Centralized):
```go
func (r *Repo) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return "", apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "get",
        })
    }
    return val, nil
}
```

**Lines saved:** 10+ lines per method! ✅

## Advanced: Pipeline Errors

```go
func (r *Repo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
    pipe := r.client.Pipeline()
    
    cmds := make([]*redis.StringCmd, len(keys))
    for i, key := range keys {
        cmds[i] = pipe.Get(ctx, key)
    }
    
    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, apperrors.HandleRedisError(ctx, err, "pipeline", map[string]any{
            "operation":  "get_multiple",
            "key_count":  len(keys),
        })
    }
    
    result := make(map[string]string)
    for i, cmd := range cmds {
        val, err := cmd.Result()
        if err == redis.Nil {
            continue // Skip missing keys
        }
        if err != nil {
            return nil, apperrors.HandleRedisError(ctx, err, keys[i], map[string]any{
                "operation": "pipeline_result",
                "index":     i,
            })
        }
        result[keys[i]] = val
    }
    
    return result, nil
}
```

## Testing

```go
func TestHandleRedisError(t *testing.T) {
    ctx := context.Background()
    
    // Test redis.Nil
    err := apperrors.HandleRedisError(ctx, redis.Nil, "user:123", nil)
    assert.Equal(t, apperrors.ErrRepoNotFound, err.Code)
    
    // Test timeout
    timeoutErr := errors.New("i/o timeout")
    err = apperrors.HandleRedisError(ctx, timeoutErr, "user:123", nil)
    assert.Equal(t, apperrors.ErrRepoTimeout, err.Code)
    
    // Test connection error
    connErr := errors.New("connection refused")
    err = apperrors.HandleRedisError(ctx, connErr, "user:123", nil)
    assert.Equal(t, apperrors.ErrRepoConnection, err.Code)
}
```

## Summary

Redis error handling is now:
- 🎯 **Centralized** - One function handles all errors
- ✅ **Automatic** - Detects error types from messages
- 📊 **Structured** - Returns proper AppError with codes
- 🚀 **Clean** - No logging in repository layer
- 💡 **Simple** - Just call `HandleRedisError()`

**Production-ready!** 🚀
