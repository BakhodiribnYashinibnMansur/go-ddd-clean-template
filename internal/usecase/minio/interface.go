package minio

import "io"

type Interface interface {
	UploadImage(imageFile io.Reader, imageSize int64, contextType string) (string, error)
	GetImageLink(imageName string) (string, error)
	UploadDoc(docFile io.Reader, docSize int64, contextType string) (string, error)
	UploadPDF(pdfFile io.Reader, pdfSize int64, contextType string) (string, error)
	DeleteFile(fileName string) error
	UploadVideo(videoFile io.Reader, videoSize int64, contextType string) (string, error)
}
