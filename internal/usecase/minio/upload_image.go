package minio

import (
	"bytes"
	"context"
	"io"

	apperrors "gct/pkg/errors"

	"github.com/disintegration/imaging"
	"go.opentelemetry.io/otel"
)

func (m *UseCase) UploadImage(ctx context.Context, imageFile io.Reader, imageSize int64, contentType string) (string, error) {
	ctx, span := otel.Tracer("minio-usecase").Start(ctx, "UploadImage")
	defer span.End()
	// Decode image
	img, err := imaging.Decode(imageFile)
	if err != nil {
		return "", apperrors.WrapServiceError(err,
			apperrors.ErrServiceInvalidInput, "failed to decode image").
			WithInput(map[string]any{"input": imageFile, "size": imageSize, "contentType": contentType})
	}

	// Encode to JPEG (CGO-free) instead of WebP
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
		return "", apperrors.WrapServiceError(err,
			apperrors.ErrServiceUnknown, "failed to encode image")
	}

	// Upload as JPEG
	// m.logger.Infow("upload image started", "size", imageSize, "contentType", contentType)

	filename, err := m.repo.Persistent.MinIO.UploadImage(ctx, &buf, int64(buf.Len()), "image/jpeg")
	if err != nil {
		// m.logger.Errorw("upload image failed", "error", err)
		return "", apperrors.MapRepoToServiceError(err).
			WithInput(map[string]any{"size": imageSize, "contentType": contentType})
	}

	// m.logger.Infow("upload image success", "filename", filename)
	return filename, nil
}
