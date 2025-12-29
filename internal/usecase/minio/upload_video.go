package minio

import (
	"context"
	"io"

	apperrors "gct/pkg/errors"
)

func (uc *UseCase) UploadVideo(videoFile io.Reader, videoSize int64, contentType string) (string, error) {
	ctx := context.Background()
	videoName, err := uc.repo.Persistent.MinIO.UploadVideo(ctx, videoFile, videoSize, contentType)
	if err != nil {
		return "", apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"input": videoFile, "size": videoSize, "contentType": contentType})
	}
	return videoName, nil
}
