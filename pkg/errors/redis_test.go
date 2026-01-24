package errors

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestHandleRedisError_NoError(t *testing.T) {
	result := HandleRedisError(nil, "user:123", nil)

	if result != nil {
		t.Errorf("HandleRedisError(nil) should return nil, got %v", result)
	}
}

func TestHandleRedisError_Nil(t *testing.T) {
	result := HandleRedisError(redis.Nil, "user:123", map[string]any{
		"operation": "get",
	})

	if result == nil {
		t.Fatal("HandleRedisError(redis.Nil) should return AppError")
	}

	if result.Type != ErrRepoNotFound {
		t.Errorf("Expected type %s, got %s", ErrRepoNotFound, result.Type)
	}

	if result.Fields["key"] != "user:123" {
		t.Errorf("Expected key=user:123, got %v", result.Fields["key"])
	}

	if result.Fields["operation"] != "get" {
		t.Errorf("Expected operation=get, got %v", result.Fields["operation"])
	}
}

func TestHandleRedisError_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:6379: connection refused")

	result := HandleRedisError(err, "user:123", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for connection error")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}

	if result.Fields["error_type"] != "connection" {
		t.Errorf("Expected error_type=connection, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_Timeout(t *testing.T) {
	err := errors.New("i/o timeout")

	result := HandleRedisError(err, "session:456", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for timeout")
	}

	if result.Type != ErrRepoTimeout {
		t.Errorf("Expected type %s, got %s", ErrRepoTimeout, result.Type)
	}

	if result.Fields["error_type"] != "timeout" {
		t.Errorf("Expected error_type=timeout, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_DeadlineExceeded(t *testing.T) {
	err := errors.New("context deadline exceeded")

	result := HandleRedisError(err, "cache:data", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for deadline exceeded")
	}

	if result.Type != ErrRepoTimeout {
		t.Errorf("Expected type %s, got %s", ErrRepoTimeout, result.Type)
	}
}

func TestHandleRedisError_WrongPassword(t *testing.T) {
	err := errors.New("WRONGPASS invalid username-password pair")

	result := HandleRedisError(err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for auth error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "auth" {
		t.Errorf("Expected error_type=auth, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_WrongType(t *testing.T) {
	err := errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

	result := HandleRedisError(err, "mykey", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for type error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "wrongtype" {
		t.Errorf("Expected error_type=wrongtype, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_OutOfMemory(t *testing.T) {
	err := errors.New("OOM command not allowed when used memory > 'maxmemory'")

	result := HandleRedisError(err, "data", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for OOM")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "oom" {
		t.Errorf("Expected error_type=oom, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_ReadOnly(t *testing.T) {
	err := errors.New("READONLY You can't write against a read only replica")

	result := HandleRedisError(err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for readonly")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "readonly" {
		t.Errorf("Expected error_type=readonly, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_ClusterDown(t *testing.T) {
	err := errors.New("CLUSTERDOWN The cluster is down")

	result := HandleRedisError(err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for cluster down")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "cluster" {
		t.Errorf("Expected error_type=cluster, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_NoScript(t *testing.T) {
	err := errors.New("NOSCRIPT No matching script. Please use EVAL")

	result := HandleRedisError(err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for noscript")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "noscript" {
		t.Errorf("Expected error_type=noscript, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_NoAuth(t *testing.T) {
	err := errors.New("NOAUTH Authentication required")

	result := HandleRedisError(err, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for noauth")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["error_type"] != "noauth" {
		t.Errorf("Expected error_type=noauth, got %v", result.Fields["error_type"])
	}
}

func TestHandleRedisError_EOF(t *testing.T) {
	err := errors.New("EOF")

	result := HandleRedisError(err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for EOF")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}
}

func TestHandleRedisError_BrokenPipe(t *testing.T) {
	err := errors.New("write: broken pipe")

	result := HandleRedisError(err, "key", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for broken pipe")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}
}

func TestHandleRedisError_GenericError(t *testing.T) {
	err := errors.New("some random redis error")

	result := HandleRedisError(err, "mykey", map[string]any{
		"operation": "set",
		"ttl":       3600,
	})

	if result == nil {
		t.Fatal("HandleRedisError should return AppError for generic error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["operation"] != "set" {
		t.Errorf("Expected operation=set, got %v", result.Fields["operation"])
	}

	if result.Fields["ttl"] != 3600 {
		t.Errorf("Expected ttl=3600, got %v", result.Fields["ttl"])
	}
}

func TestHandleRedisError_EmptyKey(t *testing.T) {

	result := HandleRedisError(redis.Nil, "", nil)

	if result == nil {
		t.Fatal("HandleRedisError should return AppError")
	}

	// Should work with empty key
	if result.Type != ErrRepoNotFound {
		t.Errorf("Expected type %s, got %s", ErrRepoNotFound, result.Type)
	}
}

func TestHandleRedisError_ExtraFields(t *testing.T) {

	extraFields := map[string]any{
		"user_id":   123,
		"operation": "cache_miss",
		"ttl":       3600,
	}

	result := HandleRedisError(redis.Nil, "user:123", extraFields)

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleRedisError(tt.err, "key", nil)

			if result == nil {
				t.Fatal("HandleRedisError should return AppError")
			}

			if result.Type != tt.expectedCode {
				t.Errorf("Expected type %s, got %s", tt.expectedCode, result.Type)
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
	for range b.N {
		HandleRedisError(redis.Nil, "key", nil)
	}
}

func BenchmarkHandleRedisError_ConnectionError(b *testing.B) {
	err := errors.New("connection refused")

	b.ResetTimer()
	for range b.N {
		HandleRedisError(err, "key", nil)
	}
}

func BenchmarkHandleRedisError_Timeout(b *testing.B) {
	err := errors.New("i/o timeout")

	b.ResetTimer()
	for range b.N {
		HandleRedisError(err, "key", nil)
	}
}

func BenchmarkHandleRedisError_GenericError(b *testing.B) {
	err := errors.New("some redis error")

	b.ResetTimer()
	for range b.N {
		HandleRedisError(err, "key", nil)
	}
}
