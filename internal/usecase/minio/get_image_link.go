package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) GetImageLink(ctx context.Context, imageName string) (string, error) {
	// m.logger.WithContext(ctx).Infow("get image link started", "imageName", imageName)

	imageLink, err := m.repo.Persistent.MinIO.GetFileURL(ctx, imageName)
	if err != nil {
		// m.logger.WithContext(ctx).Errorw("get image link failed", "error", err)
		return "", apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"imageName": imageName})
	}
	// m.logger.WithContext(ctx).Infow("get image link success")
	return imageLink, nil
}
