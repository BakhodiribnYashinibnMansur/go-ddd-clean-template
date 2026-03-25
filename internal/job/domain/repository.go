package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// JobFilter carries filtering parameters for listing jobs.
type JobFilter struct {
	TaskName *string
	Status   *string
	Limit    int64
	Offset   int64
}

// JobRepository is the write-side repository for the Job aggregate.
type JobRepository interface {
	Save(ctx context.Context, entity *Job) error
	Update(ctx context.Context, entity *Job) error
	FindByID(ctx context.Context, id uuid.UUID) (*Job, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// JobView is a read-model DTO for jobs.
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

// JobReadRepository is the read-side repository returning projected views.
type JobReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*JobView, error)
	List(ctx context.Context, filter JobFilter) ([]*JobView, int64, error)
}
