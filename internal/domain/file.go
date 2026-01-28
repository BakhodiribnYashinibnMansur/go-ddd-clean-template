package domain

import "time"

// FileInfo represents uploaded file metadata
type FileInfo struct {
	FileName    string    `json:"file_name"`
	FileURL     string    `json:"file_url,omitempty"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
	BucketName  string    `json:"bucket_name,omitempty"`
}

// FileUploadRequest represents file upload request
type FileUploadRequest struct {
	File        any // io.Reader
	FileName    string
	FileSize    int64
	ContentType string
}

// Content Type Constants
const (
	ContentTypeJPG  = "image/jpg"
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeSVG  = "image/svg+xml"
	ContentTypeHEIC = "image/heic"
	ContentTypeHEIF = "image/heif"
	ContentTypePDF  = "application/pdf"
	ContentTypeDOC  = "application/msword"
	ContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ContentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ContentTypeMP4  = "video/mp4"
	ContentTypeMPEG = "video/mpeg"
	ContentTypeAVI  = "video/x-msvideo"
	ContentType3GP  = "video/3gpp"
	ContentTypeWEBM = "video/webm"
)

// IsImageContentType checks if content type is image
func IsImageContentType(contentType string) bool {
	imageTypes := map[string]bool{
		ContentTypeJPG:  true,
		ContentTypeJPEG: true,
		ContentTypePNG:  true,
		ContentTypeSVG:  true,
		ContentTypeHEIC: true,
		ContentTypeHEIF: true,
	}
	return imageTypes[contentType]
}

// IsVideoContentType checks if content type is video
func IsVideoContentType(contentType string) bool {
	videoTypes := map[string]bool{
		ContentTypeMP4:  true,
		ContentTypeMPEG: true,
		ContentTypeAVI:  true,
		ContentType3GP:  true,
		ContentTypeWEBM: true,
	}
	return videoTypes[contentType]
}

// IsDocumentContentType checks if content type is document
func IsDocumentContentType(contentType string) bool {
	docTypes := map[string]bool{
		ContentTypePDF:  true,
		ContentTypeDOC:  true,
		ContentTypeDOCX: true,
		ContentTypeXLSX: true,
	}
	return docTypes[contentType]
}
