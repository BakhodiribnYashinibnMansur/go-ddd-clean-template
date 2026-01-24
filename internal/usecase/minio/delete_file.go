package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) DeleteFile(ctx context.Context, fileName string) error {
	// m.logger.WithContext(ctx).Infow("delete file started", "fileName", fileName)

	err := m.repo.Persistent.MinIO.DeleteFile(ctx, fileName)
	if err != nil {
		// m.logger.WithContext(ctx).Errorw("delete file failed", "error", err)
		return apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"fileName": fileName})
	}
	// m.logger.WithContext(ctx).Infow("delete file success")
	return nil
}
