package minio

import (
	"context"
	"io"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) UploadVideo(ctx context.Context, videoFile io.Reader, videoSize int64, contentType string) (string, error) {
	videoName, err := m.repo.Persistent.MinIO.UploadVideo(ctx, videoFile, videoSize, contentType)
	if err != nil {
		return "", apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"input": videoFile, "size": videoSize, "contentType": contentType})
	}
	return videoName, nil
}
