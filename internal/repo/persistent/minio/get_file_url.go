package minio

import (
	"context"
	"net/url"
	"time"

	apperrors "gct/internal/shared/infrastructure/errors"
)

// GetFileURL generates a presigned URL for a file
func (r *Repo) GetFileURL(ctx context.Context, fileName string) (string, error) {
	expiry := time.Second * 24 * 60 * 60 * 7 // 7 days

	presignedURL, err := r.client.PresignedGetObject(ctx, r.config.Bucket, fileName, expiry, url.Values{})
	if err != nil {
		return "", apperrors.HandleMinioError(err, map[string]any{"filename": fileName})
	}
	return presignedURL.String(), nil
}
