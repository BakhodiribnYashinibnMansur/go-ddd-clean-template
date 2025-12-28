package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestHandleRedisError_NoError(t *testing.T) {
	ctx := context.Background()
	result := HandleRedisError(ctx, nil, "user:123", nil)

	if result != nil {
		t.Errorf("HandleRedisError(nil) should return nil, got %v", result)
	}
}

func TestHandleRedisError_Nil(t *testing.T) {
	ctx := context.Background()
	result := HandleRedisError(ctx, redis.Nil, "user:123", map[string]any{
		"operation": "get",
	})

	if result == nil {
		t.Fatal("HandleRedisError(redis.Nil) should return AppError")
	}

	if result.Code != ErrRepoNotFound {
		t.Errorf("Expected code %s, got %s", ErrRepoNotFound, result.Code)
	}

	if result.Fields["key"] != "user:123" {
		t.Errorf("Expected key=user:123, got %v", result.Fields["key"])
	}

	if result.Fields["operation"] != "get" {
		t.Errorf("Expected operation=get, got %v", result.Fields["operation"])
	}
}

func TestHandleRedisError_ConnectionRefused(t *testing.T) {
	ctx := context.Background()
	err := errors.New("dial tcp 127.0.0.1:6379: connection refused")

	result := HandleRedisError(ctx, err, "user:123", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for connection error")
	}

	if result.Code != ErrRepoConnection {
		t.Errorf("Expected code %s, got %s", ErrRepoConnection, result.Code)
	}

	if result.Fields["error_type"] != "connection" {
		t.Errorf("Expected error_type=connection, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_Timeout(t *testing.T) {
	ctx := context.Background()
	err := errors.New("i/o timeout")

	result := HandleRedisError(ctx, err, "session:456", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for timeout")
	}

	if result.Code != ErrRepoTimeout {
		t.Errorf("Expected code %s, got %s", ErrRepoTimeout, result.Code)
	}

	if result.Fields["error_type"] != "timeout" {
		t.Errorf("Expected error_type=timeout, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_DeadlineExceeded(t *testing.T) {
	ctx := context.Background()
	err := errors.New("context deadline exceeded")

	result := HandleRedisError(ctx, err, "cache:data", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for deadline exceeded")
	}

	if result.Code != ErrRepoTimeout {
		t.Errorf("Expected code %s, got %s", ErrRepoTimeout, result.Code)
	}
}

func TestHandleRedisError_WrongPassword(t *testing.T) {
	ctx := context.Background()
	err := errors.New("WRONGPASS invalid username-password pair")

	result := HandleRedisError(ctx, err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for auth error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "auth" {
		t.Errorf("Expected error_type=auth, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_WrongType(t *testing.T) {
	ctx := context.Background()
	err := errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

	result := HandleRedisError(ctx, err, "mykey", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for type error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "wrongtype" {
		t.Errorf("Expected error_type=wrongtype, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_OutOfMemory(t *testing.T) {
	ctx := context.Background()
	err := errors.New("OOM command not allowed when used memory > 'maxmemory'")

	result := HandleRedisError(ctx, err, "data", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for OOM")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "oom" {
		t.Errorf("Expected error_type=oom, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_ReadOnly(t *testing.T) {
	ctx := context.Background()
	err := errors.New("READONLY You can't write against a read only replica")

	result := HandleRedisError(ctx, err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for readonly")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "readonly" {
		t.Errorf("Expected error_type=readonly, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_ClusterDown(t *testing.T) {
	ctx := context.Background()
	err := errors.New("CLUSTERDOWN The cluster is down")

	result := HandleRedisError(ctx, err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for cluster down")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "cluster" {
		t.Errorf("Expected error_type=cluster, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_NoScript(t *testing.T) {
	ctx := context.Background()
	err := errors.New("NOSCRIPT No matching script. Please use EVAL")

	result := HandleRedisError(ctx, err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for noscript")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "noscript" {
		t.Errorf("Expected error_type=noscript, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_NoAuth(t *testing.T) {
	ctx := context.Background()
	err := errors.New("NOAUTH Authentication required")

	result := HandleRedisError(ctx, err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for noauth")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["error_type"] != "noauth" {
		t.Errorf("Expected error_type=noauth, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_EOF(t *testing.T) {
	ctx := context.Background()
	err := errors.New("EOF")

	result := HandleRedisError(ctx, err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for EOF")
	}

	if result.Code != ErrRepoConnection {
		t.Errorf("Expected code %s, got %s", ErrRepoConnection, result.Code)
	}
}

func TestHandleRedisError_BrokenPipe(t *testing.T) {
	ctx := context.Background()
	err := errors.New("write: broken pipe")

	result := HandleRedisError(ctx, err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for broken pipe")
	}

	if result.Code != ErrRepoConnection {
		t.Errorf("Expected code %s, got %s", ErrRepoConnection, result.Code)
	}
}

func TestHandleRedisError_GenericError(t *testing.T) {
	ctx := context.Background()
	err := errors.New("some random redis error")

	result := HandleRedisError(ctx, err, "mykey", map[string]any{
		"operation": "set",
		"ttl":       3600,
	})

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for generic error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["operation"] != "set" {
		t.Errorf("Expected operation=set, got %v", result.Fields["operation"])
	}

	if result.Fields["ttl"] != 3600 {
		t.Errorf("Expected ttl=3600, got %v", result.Fields["ttl"])
	}
}

func TestHandleRedisError_EmptyKey(t *testing.T) {
	ctx := context.Background()

	result := HandleRedisError(ctx, redis.Nil, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError")
	}

	// Should work with empty key
	if result.Code != ErrRepoNotFound {
		t.Errorf("Expected code %s, got %s", ErrRepoNotFound, result.Code)
	}
}

func TestHandleRedisError_ExtraFields(t *testing.T) {
	ctx := context.Background()

	extraFields := map[string]any{
		"user_id":   123,
		"operation": "cache_miss",
		"ttl":       3600,
	}

	result := HandleRedisError(ctx, redis.Nil, "user:123", extraFields)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError")
	}

	// Check all extra fields are present
	for key, expectedValue := range extraFields {
		if result.Fields[key] != expectedValue {
			t.Errorf("Expected field %s=%v, got %v", key, expectedValue, result.Fields[key])
		}
	}
}

func TestHandleRedisError_MultiplePatterns(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
		expectedType string
	}{
		{
			name:         "connection timeout",
			err:          errors.New("dial tcp: i/o timeout"),
			expectedCode: ErrRepoTimeout,
			expectedType: "timeout",
		},
		{
			name:         "connection refused",
			err:          errors.New("connection refused"),
			expectedCode: ErrRepoConnection,
			expectedType: "connection",
		},
		{
			name:         "redis nil",
			err:          redis.Nil,
			expectedCode: ErrRepoNotFound,
			expectedType: "",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleRedisError(ctx, tt.err, "key", nil)

			if result == nil {
				t.Fatal("HandleRedisError should return AppError")
			}

			if result.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result.Code)
			}

			if tt.expectedType != "" {
				if result.Fields["error_type"] != tt.expectedType {
					t.Errorf("Expected error_type=%s, got %v", tt.expectedType, result.Fields["error_type"])
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkHandleRedisError_Nil(b *testing.B) {
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		HandleRedisError(ctx, redis.Nil, "key", nil)
	}
}

func BenchmarkHandleRedisError_ConnectionError(b *testing.B) {
	ctx := context.Background()
	err := errors.New("connection refused")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleRedisError(ctx, err, "key", nil)
	}
}

func BenchmarkHandleRedisError_Timeout(b *testing.B) {
	ctx := context.Background()
	err := errors.New("i/o timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleRedisError(ctx, err, "key", nil)
	}
}

func BenchmarkHandleRedisError_GenericError(b *testing.B) {
	ctx := context.Background()
	err := errors.New("some redis error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleRedisError(ctx, err, "key", nil)
	}
}
