package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"

	apperrors "gct/pkg/errors"
)

func (m *UseCase) UploadImage(imageFile io.Reader, imageSize int64, contentType string) (string, error) {
	// Decode image
	img, err := imaging.Decode(imageFile)
	if err != nil {
		return "", apperrors.WrapServiceError(context.Background(), err, apperrors.ErrServiceInvalidInput, "failed to decode image").
			WithInput(map[string]any{"input": imageFile, "size": imageSize, "contentType": contentType})
	}

	// Compress/Resize if needed (e.g. max width 2048)
	if img.Bounds().Dx() > 2048 {
		img = imaging.Resize(img, 2048, 0, imaging.Lanczos)
	}

	// Encode to WebP
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 80}); err != nil {
		return "", apperrors.WrapServiceError(context.Background(), err, apperrors.ErrServiceUnknown, "failed to encode image")
	}

	// Upload as WebP
	ctx := context.Background()
	filename, err := m.repo.Persistent.MinIO.UploadImage(ctx, &buf, int64(buf.Len()), "image/webp")
	if err != nil {
		return "", apperrors.MapRepoToServiceError(ctx, err).
			WithInput(map[string]any{"size": imageSize, "contentType": contentType})
	}

	return filename, nil
}
