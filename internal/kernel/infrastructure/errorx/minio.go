package errorx

import (
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
)

// HandleMinioError handles MinIO errors and converts them to AppError
// This centralizes all MinIO error handling logic
func HandleMinioError(err error, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		return handleMinioResponseError(minioErr, extraFields)
	}

	return handleMinioGenericError(err, extraFields)
}

func handleMinioResponseError(minioErr minio.ErrorResponse, extraFields map[string]any) *AppError {
	var appErr *AppError

	switch minioErr.Code {
	case "NoSuchKey", "NoSuchBucket", "NoSuchUpload", "NoSuchVersion", "ResourceNotFound":
		appErr = AutoSource(NewRepoError(ErrRepoNotFound, "minio resource not found"))
	case "AccessDenied":
		appErr = AutoSource(NewRepoError(ErrRepoDatabase, "minio access denied"))
	case "EntityTooLarge":
		appErr = AutoSource(NewRepoError(ErrRepoDatabase, "minio entity too large"))
	case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
		appErr = AutoSource(NewRepoError(ErrRepoAlreadyExists, "minio bucket already exists"))
	default:
		appErr = AutoSource(NewRepoError(ErrRepoDatabase, "minio operation failed"))
	}

	_ = appErr.WithField("minio_code", minioErr.Code)
	_ = appErr.WithDetails(minioErr.Message)

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}

	return appErr
}

func handleMinioGenericError(err error, extraFields map[string]any) *AppError {
	errMsg := err.Error()
	var appErr *AppError

	if strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "dial tcp") {
		appErr = AutoSource(NewRepoError(ErrRepoConnection, "minio connection error"))
	} else if strings.Contains(errMsg, "timeout") {
		appErr = AutoSource(NewRepoError(ErrRepoTimeout, "minio operation timeout"))
	} else {
		appErr = AutoSource(WrapRepoError(err, ErrRepoDatabase, "minio operation failed"))
	}

	for k, value := range extraFields {
		_ = appErr.WithField(k, value)
	}
	_ = appErr.WithDetails(errMsg)

	return appErr
}
