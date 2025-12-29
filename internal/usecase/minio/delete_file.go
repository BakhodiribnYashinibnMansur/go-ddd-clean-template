package minio

import (
	"context"

	apperrors "gct/pkg/errors"
)

func (uc *UseCase) DeleteFile(fileName string) error {
	ctx := context.Background()
	err := uc.repo.Persistent.MinIO.DeleteFile(ctx, fileName)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"fileName": fileName})
	}
	return nil
}
