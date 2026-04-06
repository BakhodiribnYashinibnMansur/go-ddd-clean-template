package repository

import (
	"context"
	"time"

	"gct/internal/context/content/generic/file/domain/entity"

	"github.com/google/uuid"
)

// FileFilter carries optional filtering parameters for listing files.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type FileFilter struct {
	Name     *string
	MimeType *string
	Limit    int64
	Offset   int64
}

// FileView is a read-model projection optimized for query responses.
// It is decoupled from the File aggregate to allow the read side to evolve independently.
type FileView struct {
	ID           entity.FileID `json:"id"`
	Name         string        `json:"name"`
	OriginalName string        `json:"original_name"`
	MimeType     string        `json:"mime_type"`
	Size         int64         `json:"size"`
	Path         string        `json:"path"`
	URL          string        `json:"url"`
	UploadedBy   *uuid.UUID    `json:"uploaded_by,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// FileReadRepository is the read-side repository returning projected views.
// Implementations must return ErrFileNotFound when FindByID yields no result.
// List returns matching views plus the total count for pagination.
type FileReadRepository interface {
	FindByID(ctx context.Context, id entity.FileID) (*FileView, error)
	List(ctx context.Context, filter FileFilter) ([]*FileView, int64, error)
}
