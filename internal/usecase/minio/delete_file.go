package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) DeleteFile(fileName string) error {
	ctx := context.Background()
	err := m.repo.Persistent.MinIO.DeleteFile(ctx, fileName)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"fileName": fileName})
	}
	return nil
}
