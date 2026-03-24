package minio

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
)

func (m *UseCase) DeleteFile(ctx context.Context, fileName string) error {
	// m.logger.Infow("delete file started", "fileName", fileName)

	err := m.repo.Persistent.MinIO.DeleteFile(ctx, fileName)
	if err != nil {
		// m.logger.Errorw("delete file failed", "error", err)
		return apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"fileName": fileName})
	}
	// m.logger.Infow("delete file success")
	return nil
}
