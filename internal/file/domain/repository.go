package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// FileFilter carries filtering parameters for listing files.
type FileFilter struct {
	Name     *string
	MimeType *string
	Limit    int64
	Offset   int64
}

// FileRepository is the write-side repository for the File aggregate.
type FileRepository interface {
	Save(ctx context.Context, entity *File) error
}

// FileView is a read-model DTO for files.
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

// FileReadRepository is the read-side repository returning projected views.
type FileReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*FileView, error)
	List(ctx context.Context, filter FileFilter) ([]*FileView, int64, error)
}
