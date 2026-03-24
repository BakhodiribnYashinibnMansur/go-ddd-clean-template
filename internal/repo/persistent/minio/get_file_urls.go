package minio

import (
	"context"
	"net/url"
	"time"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

// GetFileURLs generates presigned URLs for a list of files
func (r *Repo) GetFileURLs(ctx context.Context, files []domain.File) ([]domain.File, error) {
	expiry := time.Second * 24 * 60 * 60 * 7 // 7 days

	for i := range files {
		if len(files[i].Name) != 0 {
			presignedURL, err := r.client.PresignedGetObject(ctx, r.config.Bucket, files[i].Name, expiry, url.Values{})
			if err != nil {
				return files, apperrors.HandleMinioError(err, map[string]any{"filename": files[i].Name})
			}
			files[i].Link = presignedURL.String()
		}
	}
	return files, nil
}
