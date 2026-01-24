package minio

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/minio/minio-go/v7"
)

// DeleteFile deletes a file from the minio server
func (r *Repo) DeleteFile(ctx context.Context, fileName string) error {
	err := r.client.RemoveObject(ctx, r.config.Bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return apperrors.HandleMinioError(err, map[string]any{"filename": fileName})
	}
	return nil
}
