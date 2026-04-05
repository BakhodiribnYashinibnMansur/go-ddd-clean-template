package application

import (
	"time"

	"github.com/google/uuid"
)

// DataExportView is a read-model DTO returned by query handlers.
type DataExportView struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	DataType  string    `json:"data_type"`
	Format    string    `json:"format"`
	Status    string    `json:"status"`
	FileURL   *string   `json:"file_url,omitempty"`
	Error     *string   `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
