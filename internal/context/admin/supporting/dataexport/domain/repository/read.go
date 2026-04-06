package repository

import (
	"context"
	"time"

	"gct/internal/context/admin/supporting/dataexport/domain/entity"

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

// DataExportView is a read-model DTO for data exports.
type DataExportView struct {
	ID        entity.DataExportID `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	DataType  string              `json:"data_type"`
	Format    string              `json:"format"`
	Status    string              `json:"status"`
	FileURL   *string             `json:"file_url,omitempty"`
	Error     *string             `json:"error,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// DataExportReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected DataExportView DTOs, bypassing aggregate reconstruction for read performance.
type DataExportReadRepository interface {
	FindByID(ctx context.Context, id entity.DataExportID) (*DataExportView, error)
	List(ctx context.Context, filter DataExportFilter) ([]*DataExportView, int64, error)
}
