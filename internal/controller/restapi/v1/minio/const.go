package minio

import "errors"

var ErrInvalidFileFormat = errors.New("invalid file format")

const (
	headerContentType = "Content-Type"
	filePath          = "file-path" // used in download_file.go
	formFileName      = "file"      // used in upload_image.go
)

const (
	jpgContentType      string = "image/jpg"
	pngContentType      string = "image/png"
	jpegContentType     string = "image/jpeg"
	xlsxContentType     string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	xlsContentType      string = "application/vnd.ms-excel"
	docContentType      string = "application/msword"
	pdfContentType      string = "application/pdf"
	docxContentType     string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	svgContentType      string = "image/svg+xml"
	heicContentType     string = "image/heic"
	heifContentType     string = "image/heif"
	mp4ContentType      string = "video/mp4"
	mp3ContentType      string = "audio/mp3"
	mpgContentType      string = "video/mpeg"
	aviContentType      string = "video/x-msvideo"
	video3gpContentType string = "video/3gpp"
	webmContentType     string = "video/webm"
)

var imageContentTypes = map[string]bool{
	jpgContentType:  true,
	pngContentType:  true,
	jpegContentType: true,
	svgContentType:  true,
	heicContentType: true,
	heifContentType: true,
}

var videoContentTypes = map[string]bool{
	mp4ContentType:      true,
	mp3ContentType:      true,
	mpgContentType:      true,
	aviContentType:      true,
	video3gpContentType: true,
	webmContentType:     true,
}
