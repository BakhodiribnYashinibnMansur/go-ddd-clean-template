package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// File is the aggregate root for uploaded file metadata. It is immutable after creation.
type File struct {
	shared.AggregateRoot
	name         string
	originalName string
	mimeType     string
	size         int64
	path         string
	url          string
	uploadedBy   *uuid.UUID
}

// NewFile creates a new File aggregate and raises a FileUploaded event.
func NewFile(name, originalName, mimeType string, size int64, path, url string, uploadedBy *uuid.UUID) *File {
	f := &File{
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		originalName:  originalName,
		mimeType:      mimeType,
		size:          size,
		path:          path,
		url:           url,
		uploadedBy:    uploadedBy,
	}
	f.AddEvent(NewFileUploaded(f.ID(), name, mimeType, size))
	return f
}

// ReconstructFile rebuilds a File aggregate from persisted data.
func ReconstructFile(
	id uuid.UUID,
	createdAt time.Time,
	name, originalName, mimeType string,
	size int64,
	path, url string,
	uploadedBy *uuid.UUID,
) *File {
	return &File{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, createdAt, nil),
		name:          name,
		originalName:  originalName,
		mimeType:      mimeType,
		size:          size,
		path:          path,
		url:           url,
		uploadedBy:    uploadedBy,
	}
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (f *File) Name() string         { return f.name }
func (f *File) OriginalName() string { return f.originalName }
func (f *File) MimeType() string     { return f.mimeType }
func (f *File) Size() int64          { return f.size }
func (f *File) Path() string         { return f.path }
func (f *File) URL() string          { return f.url }
func (f *File) UploadedBy() *uuid.UUID { return f.uploadedBy }
