package minio

import (
	"context"
	"io"
	"strings"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadDocument uploads a document to the minio server
func (r *Repo) UploadDocument(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error) {
	fileName := uuid.New()

	// Determine extension based on content type
	fileExtension := "dat"
	if strings.Contains(contentType, "msword") {
		fileExtension = "doc"
	} else if strings.Contains(contentType, "wordprocessingml") {
		fileExtension = "docx"
	} else if strings.Contains(contentType, "pdf") {
		fileExtension = "pdf"
	} else if strings.Contains(contentType, "kth") { // handling specific cases if needed
		fileExtension = "kth"
	}

	docFileName := fileName.String() + "." + fileExtension
	_, err := r.client.PutObject(ctx, r.config.Bucket, docFileName, file, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", apperrors.HandleMinioError(ctx, err, map[string]any{"filename": docFileName})
	}
	return docFileName, nil
}
