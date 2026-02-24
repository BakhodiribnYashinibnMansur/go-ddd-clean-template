package domain

import (
	"encoding/json"
	"time"
)

type DataExport struct {
	ID          string          `json:"id" db:"id"`
	Type        string          `json:"type" db:"type"`
	Status      string          `json:"status" db:"status"`
	FileURL     *string         `json:"file_url" db:"file_url"`
	Filters     json.RawMessage `json:"filters" db:"filters"`
	CreatedBy   *string         `json:"created_by" db:"created_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	CompletedAt *time.Time      `json:"completed_at" db:"completed_at"`
}

type DataExportFilter struct {
	Type   string
	Status string
	Limit  int
	Offset int
}

type CreateDataExportRequest struct {
	Type    string          `json:"type" binding:"required"`
	Filters json.RawMessage `json:"filters"`
}
