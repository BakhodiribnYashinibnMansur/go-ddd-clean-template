package domain

import (
	"time"

	"github.com/google/uuid"
)

// JobScheduled is a domain event raised when a new job enters the queue.
// Subscribers can use this to trigger immediate execution or update dashboard counters.
type JobScheduled struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	TaskName    string
}

func NewJobScheduled(id uuid.UUID, taskName string) JobScheduled {
	return JobScheduled{
		aggregateID: id,
		occurredAt:  time.Now(),
		TaskName:    taskName,
	}
}

func (e JobScheduled) EventName() string      { return "job.scheduled" }
func (e JobScheduled) OccurredAt() time.Time   { return e.occurredAt }
func (e JobScheduled) AggregateID() uuid.UUID  { return e.aggregateID }

// JobCompleted is a domain event raised when a job finishes successfully.
// This is only emitted on success — failures do not produce a domain event.
type JobCompleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	TaskName    string
}

func NewJobCompleted(id uuid.UUID, taskName string) JobCompleted {
	return JobCompleted{
		aggregateID: id,
		occurredAt:  time.Now(),
		TaskName:    taskName,
	}
}

func (e JobCompleted) EventName() string      { return "job.completed" }
func (e JobCompleted) OccurredAt() time.Time   { return e.occurredAt }
func (e JobCompleted) AggregateID() uuid.UUID  { return e.aggregateID }
