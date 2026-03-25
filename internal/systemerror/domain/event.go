package domain

import (
	"time"

	"github.com/google/uuid"
)

// SystemErrorRecorded is raised when a new system error is recorded.
type SystemErrorRecorded struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Code        string
	Severity    string
}

func NewSystemErrorRecorded(id uuid.UUID, code, severity string) SystemErrorRecorded {
	return SystemErrorRecorded{
		aggregateID: id,
		occurredAt:  time.Now(),
		Code:        code,
		Severity:    severity,
	}
}

func (e SystemErrorRecorded) EventName() string      { return "system_error.recorded" }
func (e SystemErrorRecorded) OccurredAt() time.Time   { return e.occurredAt }
func (e SystemErrorRecorded) AggregateID() uuid.UUID  { return e.aggregateID }

// SystemErrorResolved is raised when a system error is resolved.
type SystemErrorResolved struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	ResolvedBy  uuid.UUID
}

func NewSystemErrorResolved(id, resolvedBy uuid.UUID) SystemErrorResolved {
	return SystemErrorResolved{
		aggregateID: id,
		occurredAt:  time.Now(),
		ResolvedBy:  resolvedBy,
	}
}

func (e SystemErrorResolved) EventName() string      { return "system_error.resolved" }
func (e SystemErrorResolved) OccurredAt() time.Time   { return e.occurredAt }
func (e SystemErrorResolved) AggregateID() uuid.UUID  { return e.aggregateID }
