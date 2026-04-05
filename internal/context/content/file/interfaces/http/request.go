package http

import "github.com/google/uuid"

// CreateRequest represents the request body for creating a file record.
type CreateRequest struct {
	Name         string     `json:"name" binding:"required"`
	OriginalName string     `json:"original_name" binding:"required"`
	MimeType     string     `json:"mime_type" binding:"required"`
	Size         int64      `json:"size" binding:"required"`
	Path         string     `json:"path" binding:"required"`
	URL          string     `json:"url" binding:"required"`
	UploadedBy   *uuid.UUID `json:"uploaded_by,omitempty"`
}
