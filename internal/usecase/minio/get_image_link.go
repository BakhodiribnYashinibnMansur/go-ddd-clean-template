package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) GetImageLink(imageName string) (string, error) {
	ctx := context.Background()
	imageLink, err := m.repo.Persistent.MinIO.GetFileURL(ctx, imageName)
	if err != nil {
		return "", apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"imageName": imageName})
	}
	return imageLink, nil
}
