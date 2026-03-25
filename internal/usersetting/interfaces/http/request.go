package http

import "github.com/google/uuid"

// UpsertRequest represents the request body for creating or updating a user setting.
type UpsertRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Key    string    `json:"key" binding:"required"`
	Value  string    `json:"value" binding:"required"`
}
