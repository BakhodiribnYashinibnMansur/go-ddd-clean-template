package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Job is the aggregate root for background/scheduled jobs.
type Job struct {
	shared.AggregateRoot
	taskName    string
	status      string
	payload     map[string]any
	result      map[string]any
	attempts    int
	maxAttempts int
	scheduledAt *time.Time
	startedAt   *time.Time
	completedAt *time.Time
	errorMsg    *string
}

// Job status constants.
const (
	JobStatusPending   = "PENDING"
	JobStatusRunning   = "RUNNING"
	JobStatusCompleted = "COMPLETED"
	JobStatusFailed    = "FAILED"
)

// NewJob creates a new Job aggregate and raises a JobScheduled event.
func NewJob(taskName string, payload map[string]any, maxAttempts int, scheduledAt *time.Time) *Job {
	if payload == nil {
		payload = make(map[string]any)
	}
	j := &Job{
		AggregateRoot: shared.NewAggregateRoot(),
		taskName:      taskName,
		status:        JobStatusPending,
		payload:       payload,
		result:        make(map[string]any),
		attempts:      0,
		maxAttempts:   maxAttempts,
		scheduledAt:   scheduledAt,
	}
	j.AddEvent(NewJobScheduled(j.ID(), taskName))
	return j
}

// ReconstructJob rebuilds a Job aggregate from persisted data.
func ReconstructJob(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	taskName, status string,
	payload, result map[string]any,
	attempts, maxAttempts int,
	scheduledAt, startedAt, completedAt *time.Time,
	errorMsg *string,
) *Job {
	if payload == nil {
		payload = make(map[string]any)
	}
	if result == nil {
		result = make(map[string]any)
	}
	return &Job{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		taskName:      taskName,
		status:        status,
		payload:       payload,
		result:        result,
		attempts:      attempts,
		maxAttempts:   maxAttempts,
		scheduledAt:   scheduledAt,
		startedAt:     startedAt,
		completedAt:   completedAt,
		errorMsg:      errorMsg,
	}
}

// Complete marks the job as completed with a result.
func (j *Job) Complete(result map[string]any) {
	now := time.Now()
	j.status = JobStatusCompleted
	j.result = result
	j.completedAt = &now
	j.Touch()
	j.AddEvent(NewJobCompleted(j.ID(), j.taskName))
}

// Fail marks the job as failed with an error message.
func (j *Job) Fail(errMsg string) {
	j.status = JobStatusFailed
	j.errorMsg = &errMsg
	j.Touch()
}

// Start marks the job as running.
func (j *Job) Start() {
	now := time.Now()
	j.status = JobStatusRunning
	j.startedAt = &now
	j.attempts++
	j.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (j *Job) TaskName() string        { return j.taskName }
func (j *Job) Status() string          { return j.status }
func (j *Job) Payload() map[string]any { return j.payload }
func (j *Job) Result() map[string]any  { return j.result }
func (j *Job) Attempts() int           { return j.attempts }
func (j *Job) MaxAttempts() int        { return j.maxAttempts }
func (j *Job) ScheduledAt() *time.Time { return j.scheduledAt }
func (j *Job) StartedAt() *time.Time   { return j.startedAt }
func (j *Job) CompletedAt() *time.Time { return j.completedAt }
func (j *Job) Error() *string          { return j.errorMsg }
