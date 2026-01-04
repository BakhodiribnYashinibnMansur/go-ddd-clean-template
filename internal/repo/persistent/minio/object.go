package minio

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/minio/minio-go/v7"
)

// ObjectExists checks if an object exists in the minio server
func (r *Repo) ObjectExists(ctx context.Context, fileName string) error {
	_, err := r.client.StatObject(ctx, r.config.Bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return apperrors.HandleMinioError(ctx, err, map[string]any{"filename": fileName})
	}
	return nil
}
