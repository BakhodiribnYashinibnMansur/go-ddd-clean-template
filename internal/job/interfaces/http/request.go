package http

import "time"

// CreateRequest represents the request body for creating a job.
type CreateRequest struct {
	TaskName    string         `json:"task_name" binding:"required"`
	Payload     map[string]any `json:"payload,omitempty"`
	MaxAttempts int            `json:"max_attempts"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
}

// UpdateRequest represents the request body for updating a job.
type UpdateRequest struct {
	Status *string        `json:"status,omitempty"`
	Result map[string]any `json:"result,omitempty"`
	Error  *string        `json:"error,omitempty"`
}
