package minio

import (
	"context"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	apperrors "gct/pkg/errors"
)

// UploadFile uploads any file to the minio server from a local path
func (r *Repo) UploadFile(ctx context.Context, filePath, contentType string) (string, error) {
	fileName := uuid.New().String() + filepath.Ext(filePath)
	_, err := r.client.FPutObject(ctx, r.config.Bucket, fileName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", apperrors.HandleMinioError(ctx, err, map[string]any{"filename": fileName})
	}
	return fileName, nil
}
