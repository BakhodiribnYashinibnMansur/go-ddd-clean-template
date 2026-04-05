package http

import "github.com/google/uuid"

// CreateRequest represents the request body for creating a data export.
type CreateRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	DataType string    `json:"data_type" binding:"required"`
	Format   string    `json:"format" binding:"required"`
}

// UpdateRequest represents the request body for updating a data export.
type UpdateRequest struct {
	Status  *string `json:"status,omitempty"`
	FileURL *string `json:"file_url,omitempty"`
	Error   *string `json:"error,omitempty"`
}
