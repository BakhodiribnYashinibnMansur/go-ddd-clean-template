package domain

import (
	"time"

	"github.com/google/uuid"
)

// ErrorCodeUpdated is raised when an error code is created or updated.
type ErrorCodeUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Code        string
	Message     string
}

func NewErrorCodeUpdated(id uuid.UUID, code, message string) ErrorCodeUpdated {
	return ErrorCodeUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
		Code:        code,
		Message:     message,
	}
}

func (e ErrorCodeUpdated) EventName() string      { return "errorcode.updated" }
func (e ErrorCodeUpdated) OccurredAt() time.Time   { return e.occurredAt }
func (e ErrorCodeUpdated) AggregateID() uuid.UUID  { return e.aggregateID }
