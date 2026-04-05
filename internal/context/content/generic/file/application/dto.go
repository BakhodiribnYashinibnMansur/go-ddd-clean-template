package application

import (
	"time"

	"github.com/google/uuid"
)

// FileView is a read-model DTO returned by query handlers.
type FileView struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	Path         string     `json:"path"`
	URL          string     `json:"url"`
	UploadedBy   *uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
