package http

import "github.com/google/uuid"

// CreateRequest represents the request body for creating a notification.
type CreateRequest struct {
	UserID  uuid.UUID `json:"user_id" binding:"required"`
	Title   string    `json:"title" binding:"required"`
	Message string    `json:"message" binding:"required"`
	Type    string    `json:"type" binding:"required"`
}
