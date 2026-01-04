package minio

import (
	"bytes"
	"context"
	"io"

	apperrors "gct/pkg/errors"
	"github.com/disintegration/imaging"
)

func (m *UseCase) UploadImage(imageFile io.Reader, imageSize int64, contentType string) (string, error) {
	// Decode image
	img, err := imaging.Decode(imageFile)
	if err != nil {
		return "", apperrors.WrapServiceError(context.Background(), err,
			apperrors.ErrServiceInvalidInput, "failed to decode image").
			WithInput(map[string]any{"input": imageFile, "size": imageSize, "contentType": contentType})
	}

	// Encode to JPEG (CGO-free) instead of WebP
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
		return "", apperrors.WrapServiceError(context.Background(), err,
			apperrors.ErrServiceUnknown, "failed to encode image")
	}

	// Upload as JPEG
	ctx := context.Background()
	filename, err := m.repo.Persistent.MinIO.UploadImage(ctx, &buf, int64(buf.Len()), "image/jpeg")
	if err != nil {
		return "", apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"size": imageSize, "contentType": contentType})
	}

	return filename, nil
}
