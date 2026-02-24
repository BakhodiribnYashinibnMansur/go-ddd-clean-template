package domain

import "time"

// FileMetadata represents a record in the file_metadata table
type FileMetadata struct {
	ID           string    `json:"id" db:"id"`
	OriginalName string    `json:"original_name" db:"original_name"`
	StoredName   string    `json:"stored_name" db:"stored_name"`
	Bucket       string    `json:"bucket" db:"bucket"`
	URL          string    `json:"url" db:"url"`
	Size         int64     `json:"size" db:"size"`
	MimeType     *string   `json:"mime_type" db:"mime_type"`
	UploadedBy   *string   `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FileMetadataFilter holds pagination and filter options for listing file_metadata
type FileMetadataFilter struct {
	Search   string
	MimeType string
	Limit    int
	Offset   int
}

// UpdateFileMetadataRequest is the request body for updating file metadata
type UpdateFileMetadataRequest struct {
	OriginalName *string `json:"original_name"`
}

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
