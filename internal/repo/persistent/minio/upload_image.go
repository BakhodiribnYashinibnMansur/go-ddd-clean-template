package minio

import (
	"context"
	"io"
	"strings"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadImage uploads an image to the minio server
func (r *Repo) UploadImage(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error) {
	fileName := uuid.New()
	fileExtension := strings.Split(contentType, "/")[1]
	if contentType == "image/svg+xml" {
		fileExtension = "svg"
	}
	imageName := fileName.String() + "." + fileExtension

	_, err := r.client.PutObject(ctx, r.config.Bucket, imageName, file, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", apperrors.HandleMinioError(err, map[string]any{"filename": imageName})
	}
	return imageName, nil
}
