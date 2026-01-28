package minio

import (
	"context"
	"io"
)

type Interface interface {
	UploadImage(ctx context.Context, imageFile io.Reader, imageSize int64, contextType string) (string, error)
	GetImageLink(ctx context.Context, imageName string) (string, error)
	UploadDoc(ctx context.Context, docFile io.Reader, docSize int64, contextType string) (string, error)
	UploadPDF(ctx context.Context, pdfFile io.Reader, pdfSize int64, contextType string) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
	UploadVideo(ctx context.Context, videoFile io.Reader, videoSize int64, contextType string) (string, error)
}
