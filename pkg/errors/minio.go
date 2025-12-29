package errors

import (
	"context"
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
)

// HandleMinioError handles MinIO errors and converts them to AppError
// This centralizes all MinIO error handling logic
func HandleMinioError(ctx context.Context, err error, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		return handleMinioResponseError(ctx, minioErr, extraFields)
	}

	return handleMinioGenericError(ctx, err, extraFields)
}

func handleMinioResponseError(ctx context.Context, minioErr minio.ErrorResponse, extraFields map[string]any) *AppError {
	var appErr *AppError

	switch minioErr.Code {
	case "NoSuchKey", "NoSuchBucket", "NoSuchUpload", "NoSuchVersion", "ResourceNotFound":
		appErr = AutoSource(NewRepoError(ctx, ErrRepoNotFound, "minio resource not found"))
	case "AccessDenied":
		appErr = AutoSource(NewRepoError(ctx, ErrRepoDatabase, "minio access denied"))
	case "EntityTooLarge":
		appErr = AutoSource(NewRepoError(ctx, ErrRepoDatabase, "minio entity too large"))
	case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
		appErr = AutoSource(NewRepoError(ctx, ErrRepoAlreadyExists, "minio bucket already exists"))
	default:
		appErr = AutoSource(NewRepoError(ctx, ErrRepoDatabase, "minio operation failed"))
	}

	_ = appErr.WithField("minio_code", minioErr.Code)
	_ = appErr.WithDetails(minioErr.Message)

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}

	return appErr
}

func handleMinioGenericError(ctx context.Context, err error, extraFields map[string]any) *AppError {
	errMsg := err.Error()
	var appErr *AppError

	if strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "dial tcp") {
		appErr = AutoSource(NewRepoError(ctx, ErrRepoConnection, "minio connection error"))
	} else if strings.Contains(errMsg, "timeout") {
		appErr = AutoSource(NewRepoError(ctx, ErrRepoTimeout, "minio operation timeout"))
	} else {
		appErr = AutoSource(WrapRepoError(ctx, err, ErrRepoDatabase, "minio operation failed"))
	}

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	_ = appErr.WithDetails(errMsg)

	return appErr
}
