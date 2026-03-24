package minio

import (
	"context"
	"io"

	apperrors "gct/internal/shared/infrastructure/errors"
)

func (m *UseCase) UploadPDF(ctx context.Context, docFile io.Reader, docSize int64, contentType string) (string, error) {
	docName, err := m.repo.Persistent.MinIO.UploadDocument(ctx, docFile, docSize, contentType)
	if err != nil {
		return "", apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"input": docFile, "size": docSize, "contentType": contentType})
	}
	return docName, nil
}
