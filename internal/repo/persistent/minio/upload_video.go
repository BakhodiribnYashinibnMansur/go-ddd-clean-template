package minio

import (
	"context"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	apperrors "gct/pkg/errors"
)

// UploadVideo uploads a video to the minio server
func (r *Repo) UploadVideo(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error) {
	fileName := uuid.New()
	fileExtension := "mp4" // simple default, ideally map contentType
	if strings.Contains(contentType, "webm") {
		fileExtension = "webm"
	}

	videoFileName := fileName.String() + "." + fileExtension
	_, err := r.client.PutObject(ctx, r.config.Bucket, videoFileName, file, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", apperrors.HandleMinioError(ctx, err, map[string]any{"filename": videoFileName})
	}
	return videoFileName, nil
}
