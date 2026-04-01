package errors

import (
	"errors"
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestHandleMinioError_Nil(t *testing.T) {
	result := HandleMinioError(nil, nil)
	if result != nil {
		t.Error("expected nil for nil error input")
	}
}

func TestHandleMinioError_MinioResponseErrors(t *testing.T) {
	tests := []struct {
		name      string
		minioCode string
		wantType  string
		wantMsg   string
	}{
		{
			name:      "NoSuchKey",
			minioCode: "NoSuchKey",
			wantType:  ErrRepoNotFound,
			wantMsg:   "minio resource not found",
		},
		{
			name:      "NoSuchBucket",
			minioCode: "NoSuchBucket",
			wantType:  ErrRepoNotFound,
			wantMsg:   "minio resource not found",
		},
		{
			name:      "ResourceNotFound",
			minioCode: "ResourceNotFound",
			wantType:  ErrRepoNotFound,
			wantMsg:   "minio resource not found",
		},
		{
			name:      "AccessDenied",
			minioCode: "AccessDenied",
			wantType:  ErrRepoDatabase,
			wantMsg:   "minio access denied",
		},
		{
			name:      "EntityTooLarge",
			minioCode: "EntityTooLarge",
			wantType:  ErrRepoDatabase,
			wantMsg:   "minio entity too large",
		},
		{
			name:      "BucketAlreadyExists",
			minioCode: "BucketAlreadyExists",
			wantType:  ErrRepoAlreadyExists,
			wantMsg:   "minio bucket already exists",
		},
		{
			name:      "BucketAlreadyOwnedByYou",
			minioCode: "BucketAlreadyOwnedByYou",
			wantType:  ErrRepoAlreadyExists,
			wantMsg:   "minio bucket already exists",
		},
		{
			name:      "unknown minio code",
			minioCode: "SomeOtherError",
			wantType:  ErrRepoDatabase,
			wantMsg:   "minio operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minioErr := minio.ErrorResponse{
				Code:    tt.minioCode,
				Message: "test minio message",
			}

			result := HandleMinioError(minioErr, nil)

			if result == nil {
				t.Fatal("expected non-nil error")
			}
			if result.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, result.Type)
			}
			if result.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, result.Message)
			}
			if result.Fields["minio_code"] != tt.minioCode {
				t.Errorf("expected minio_code %q, got %v", tt.minioCode, result.Fields["minio_code"])
			}
		})
	}
}

func TestHandleMinioError_WithExtraFields(t *testing.T) {
	minioErr := minio.ErrorResponse{
		Code:    "NoSuchKey",
		Message: "key not found",
	}

	extra := map[string]any{
		"bucket": "my-bucket",
		"key":    "my-key",
	}

	result := HandleMinioError(minioErr, extra)

	if result == nil {
		t.Fatal("expected non-nil error")
	}
	if result.Fields["bucket"] != "my-bucket" {
		t.Errorf("expected bucket 'my-bucket', got %v", result.Fields["bucket"])
	}
	if result.Fields["key"] != "my-key" {
		t.Errorf("expected key 'my-key', got %v", result.Fields["key"])
	}
}

func TestHandleMinioError_GenericErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		wantType string
	}{
		{
			name:     "connection error",
			errMsg:   "connection refused",
			wantType: ErrRepoConnection,
		},
		{
			name:     "dial tcp error",
			errMsg:   "dial tcp 127.0.0.1:9000: connection refused",
			wantType: ErrRepoConnection,
		},
		{
			name:     "timeout error",
			errMsg:   "request timeout",
			wantType: ErrRepoTimeout,
		},
		{
			name:     "generic error",
			errMsg:   "something unexpected",
			wantType: ErrRepoDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := HandleMinioError(err, nil)

			if result == nil {
				t.Fatal("expected non-nil error")
			}
			if result.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, result.Type)
			}
		})
	}
}

func TestHandleMinioError_GenericWithExtraFields(t *testing.T) {
	err := errors.New("something failed")
	extra := map[string]any{"op": "upload"}

	result := HandleMinioError(err, extra)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
	if result.Fields["op"] != "upload" {
		t.Errorf("expected op 'upload', got %v", result.Fields["op"])
	}
}
