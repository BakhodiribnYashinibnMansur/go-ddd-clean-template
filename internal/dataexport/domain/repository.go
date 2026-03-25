package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DataExportFilter carries filtering parameters for listing data exports.
type DataExportFilter struct {
	UserID   *uuid.UUID
	DataType *string
	Status   *string
	Limit    int64
	Offset   int64
}

// DataExportRepository is the write-side repository for the DataExport aggregate.
type DataExportRepository interface {
	Save(ctx context.Context, entity *DataExport) error
	Update(ctx context.Context, entity *DataExport) error
	FindByID(ctx context.Context, id uuid.UUID) (*DataExport, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// DataExportView is a read-model DTO for data exports.
type DataExportView struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	DataType  string     `json:"data_type"`
	Format    string     `json:"format"`
	Status    string     `json:"status"`
	FileURL   *string    `json:"file_url,omitempty"`
	Error     *string    `json:"error,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// DataExportReadRepository is the read-side repository returning projected views.
type DataExportReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*DataExportView, error)
	List(ctx context.Context, filter DataExportFilter) ([]*DataExportView, int64, error)
}
