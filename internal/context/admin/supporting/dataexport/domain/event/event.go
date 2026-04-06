package event

import (
	"time"

	"github.com/google/uuid"
)

// ExportRequested is a domain event emitted when a user initiates a data export.
// A background job handler should subscribe to this event to begin async processing.
type ExportRequested struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	UserID      uuid.UUID
	DataType    string
}

func NewExportRequested(id, userID uuid.UUID, dataType string) ExportRequested {
	return ExportRequested{
		aggregateID: id,
		occurredAt:  time.Now(),
		UserID:      userID,
		DataType:    dataType,
	}
}

func (e ExportRequested) EventName() string      { return "dataexport.requested" }
func (e ExportRequested) OccurredAt() time.Time  { return e.occurredAt }
func (e ExportRequested) AggregateID() uuid.UUID { return e.aggregateID }

// ExportCompleted is emitted when the export file has been generated and uploaded.
// Consumers can use this to send download-ready notifications to the requesting user.
type ExportCompleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	UserID      uuid.UUID
	FileURL     string
}

func NewExportCompleted(id, userID uuid.UUID, fileURL string) ExportCompleted {
	return ExportCompleted{
		aggregateID: id,
		occurredAt:  time.Now(),
		UserID:      userID,
		FileURL:     fileURL,
	}
}

func (e ExportCompleted) EventName() string      { return "dataexport.completed" }
func (e ExportCompleted) OccurredAt() time.Time  { return e.occurredAt }
func (e ExportCompleted) AggregateID() uuid.UUID { return e.aggregateID }
