package http

import (
	"time"

	shared "gct/internal/shared/domain"
)

// CreateRequest represents the request body for creating an announcement.
type CreateRequest struct {
	Title     shared.Lang `json:"title" binding:"required"`
	Content   shared.Lang `json:"content" binding:"required"`
	Priority  int         `json:"priority"`
	StartDate *time.Time  `json:"start_date,omitempty"`
	EndDate   *time.Time  `json:"end_date,omitempty"`
}

// UpdateRequest represents the request body for updating an announcement.
type UpdateRequest struct {
	Title     *shared.Lang `json:"title,omitempty"`
	Content   *shared.Lang `json:"content,omitempty"`
	Priority  *int         `json:"priority,omitempty"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Publish   bool         `json:"publish"`
}
