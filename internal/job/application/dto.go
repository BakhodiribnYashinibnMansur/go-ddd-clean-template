package application

import (
	"time"

	"github.com/google/uuid"
)

// JobView is a read-model DTO returned by query handlers.
type JobView struct {
	ID          uuid.UUID      `json:"id"`
	TaskName    string         `json:"task_name"`
	Status      string         `json:"status"`
	Payload     map[string]any `json:"payload"`
	Result      map[string]any `json:"result"`
	Attempts    int            `json:"attempts"`
	MaxAttempts int            `json:"max_attempts"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Error       *string        `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
