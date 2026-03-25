package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExportRequested is raised when a new data export is requested.
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
func (e ExportRequested) OccurredAt() time.Time   { return e.occurredAt }
func (e ExportRequested) AggregateID() uuid.UUID  { return e.aggregateID }

// ExportCompleted is raised when a data export finishes successfully.
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
func (e ExportCompleted) OccurredAt() time.Time   { return e.occurredAt }
func (e ExportCompleted) AggregateID() uuid.UUID  { return e.aggregateID }
