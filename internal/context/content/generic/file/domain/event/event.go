package event

import (
	"time"

	"github.com/google/uuid"
)

// FileUploaded is a domain event raised when a new file is persisted.
// Downstream subscribers can use this to trigger virus scanning, thumbnail generation, or audit logging.
type FileUploaded struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Name        string
	MimeType    string
	Size        int64
}

func NewFileUploaded(id uuid.UUID, name, mimeType string, size int64) FileUploaded {
	return FileUploaded{
		aggregateID: id,
		occurredAt:  time.Now(),
		Name:        name,
		MimeType:    mimeType,
		Size:        size,
	}
}

func (e FileUploaded) EventName() string     { return "file.uploaded" }
func (e FileUploaded) OccurredAt() time.Time  { return e.occurredAt }
func (e FileUploaded) AggregateID() uuid.UUID { return e.aggregateID }
