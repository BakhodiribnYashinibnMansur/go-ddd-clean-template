package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DataExportFilter carries optional criteria for querying data exports.
// Nil pointer fields are treated as "no filter" by repository implementations.
type DataExportFilter struct {
	UserID   *uuid.UUID
	DataType *string
	Status   *string
	Limit    int64
	Offset   int64
}

// DataExportRepository is the write-side repository for the DataExport aggregate.
// Implementations must return ErrDataExportNotFound from FindByID when no row matches.
// Update must persist status transitions and file URL changes atomically.
type DataExportRepository interface {
	Save(ctx context.Context, entity *DataExport) error
	Update(ctx context.Context, entity *DataExport) error
	FindByID(ctx context.Context, id DataExportID) (*DataExport, error)
	Delete(ctx context.Context, id DataExportID) error
}

// DataExportView is a read-model DTO for data exports.
type DataExportView struct {
	ID        DataExportID `json:"id"`
	UserID    uuid.UUID    `json:"user_id"`
	DataType  string       `json:"data_type"`
	Format    string       `json:"format"`
	Status    string       `json:"status"`
	FileURL   *string      `json:"file_url,omitempty"`
	Error     *string      `json:"error,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// DataExportReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected DataExportView DTOs, bypassing aggregate reconstruction for read performance.
type DataExportReadRepository interface {
	FindByID(ctx context.Context, id DataExportID) (*DataExportView, error)
	List(ctx context.Context, filter DataExportFilter) ([]*DataExportView, int64, error)
}
