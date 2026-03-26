package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// DataExport is the aggregate root for data export requests.
// It models a state machine: PENDING -> PROCESSING -> COMPLETED | FAILED.
// Transitions are enforced by the Complete, Fail, and StartProcessing methods.
// The fileURL is only populated on successful completion.
type DataExport struct {
	shared.AggregateRoot
	userID   uuid.UUID
	dataType string
	format   string
	status   string
	fileURL  *string
	errorMsg *string
}

// DataExport status constants representing the export lifecycle state machine.
// Transitions: PENDING -> PROCESSING -> COMPLETED or FAILED. No backward transitions are allowed.
const (
	ExportStatusPending    = "PENDING"
	ExportStatusProcessing = "PROCESSING"
	ExportStatusCompleted  = "COMPLETED"
	ExportStatusFailed     = "FAILED"
)

// NewDataExport creates a new DataExport aggregate and raises an ExportRequested event.
func NewDataExport(userID uuid.UUID, dataType, format string) *DataExport {
	de := &DataExport{
		AggregateRoot: shared.NewAggregateRoot(),
		userID:        userID,
		dataType:      dataType,
		format:        format,
		status:        ExportStatusPending,
	}
	de.AddEvent(NewExportRequested(de.ID(), userID, dataType))
	return de
}

// ReconstructDataExport rebuilds a DataExport aggregate from persisted data.
func ReconstructDataExport(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	userID uuid.UUID,
	dataType, format, status string,
	fileURL, errorMsg *string,
) *DataExport {
	return &DataExport{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		userID:        userID,
		dataType:      dataType,
		format:        format,
		status:        status,
		fileURL:       fileURL,
		errorMsg:      errorMsg,
	}
}

// Complete transitions the export to COMPLETED and stores the download URL.
// An ExportCompleted event is raised for downstream notification handlers.
func (de *DataExport) Complete(fileURL string) {
	de.status = ExportStatusCompleted
	de.fileURL = &fileURL
	de.Touch()
	de.AddEvent(NewExportCompleted(de.ID(), de.userID, fileURL))
}

// Fail transitions the export to FAILED and records the error message.
// No domain event is raised on failure — callers should log the error externally if needed.
func (de *DataExport) Fail(errMsg string) {
	de.status = ExportStatusFailed
	de.errorMsg = &errMsg
	de.Touch()
}

// StartProcessing transitions the export from PENDING to PROCESSING.
// This should be called by the background worker before beginning the actual export job.
func (de *DataExport) StartProcessing() {
	de.status = ExportStatusProcessing
	de.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (de *DataExport) UserID() uuid.UUID { return de.userID }
func (de *DataExport) DataType() string  { return de.dataType }
func (de *DataExport) Format() string    { return de.format }
func (de *DataExport) Status() string    { return de.status }
func (de *DataExport) FileURL() *string  { return de.fileURL }
func (de *DataExport) Error() *string    { return de.errorMsg }
