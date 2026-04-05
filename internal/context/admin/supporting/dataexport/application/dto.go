package application

import (
	"time"

	"gct/internal/context/admin/supporting/dataexport/domain"

	"github.com/google/uuid"
)

// DataExportView is a read-model DTO returned by query handlers.
type DataExportView struct {
	ID        domain.DataExportID `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	DataType  string              `json:"data_type"`
	Format    string              `json:"format"`
	Status    string              `json:"status"`
	FileURL   *string             `json:"file_url,omitempty"`
	Error     *string             `json:"error,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}
