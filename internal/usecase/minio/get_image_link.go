package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) GetImageLink(ctx context.Context, imageName string) (string, error) {
	// m.logger.Infow("get image link started", "imageName", imageName)

	imageLink, err := m.repo.Persistent.MinIO.GetFileURL(ctx, imageName)
	if err != nil {
		// m.logger.Errorw("get image link failed", "error", err)
		return "", apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"imageName": imageName})
	}
	// m.logger.Infow("get image link success")
	return imageLink, nil
}
