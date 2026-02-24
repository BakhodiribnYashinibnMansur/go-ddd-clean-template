package domain

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	CronSchedule string         `json:"cron_schedule"`
	Payload      map[string]any `json:"payload"`
	IsActive     bool           `json:"is_active"`
	Status       string         `json:"status"` // idle, running, done, failed
	LastRunAt    *time.Time     `json:"last_run_at,omitempty"`
	NextRunAt    *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type JobFilter struct {
	Search   string
	Status   string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateJobRequest struct {
	Name         string         `json:"name" binding:"required,min=2,max=100"`
	Type         string         `json:"type" binding:"required"`
	CronSchedule string         `json:"cron_schedule"`
	Payload      map[string]any `json:"payload"`
	IsActive     bool           `json:"is_active"`
}

type UpdateJobRequest struct {
	Name         *string        `json:"name" binding:"omitempty,min=2,max=100"`
	Type         *string        `json:"type"`
	CronSchedule *string        `json:"cron_schedule"`
	Payload      map[string]any `json:"payload"`
	IsActive     *bool          `json:"is_active"`
}
