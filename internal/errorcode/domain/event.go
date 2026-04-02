package domain

import (
	"time"

	"github.com/google/uuid"
)

// ErrorCodeCreated is a domain event emitted when a new error code is created.
// Consumers can use this to add the code to a cached error-code lookup table.
type ErrorCodeCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Code        string
	Message     string
	HTTPStatus  int
}

func NewErrorCodeCreated(id uuid.UUID, code, message string, httpStatus int) ErrorCodeCreated {
	return ErrorCodeCreated{
		aggregateID: id,
		occurredAt:  time.Now(),
		Code:        code,
		Message:     message,
		HTTPStatus:  httpStatus,
	}
}

func (e ErrorCodeCreated) EventName() string     { return "errorcode.created" }
func (e ErrorCodeCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e ErrorCodeCreated) AggregateID() uuid.UUID { return e.aggregateID }

// ErrorCodeUpdated is a domain event emitted on modification of an error code.
// Consumers can use this to refresh a cached error-code lookup table used by API error mappers.
type ErrorCodeUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Code        string
	Message     string
	HTTPStatus  int
}

func NewErrorCodeUpdated(id uuid.UUID, code, message string, httpStatus int) ErrorCodeUpdated {
	return ErrorCodeUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
		Code:        code,
		Message:     message,
		HTTPStatus:  httpStatus,
	}
}

func (e ErrorCodeUpdated) EventName() string     { return "errorcode.updated" }
func (e ErrorCodeUpdated) OccurredAt() time.Time  { return e.occurredAt }
func (e ErrorCodeUpdated) AggregateID() uuid.UUID { return e.aggregateID }

// ErrorCodeDeleted is a domain event emitted when an error code is removed.
// Consumers should use this to evict the code from any in-memory lookup caches.
type ErrorCodeDeleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Code        string
}

func NewErrorCodeDeleted(id uuid.UUID, code string) ErrorCodeDeleted {
	return ErrorCodeDeleted{
		aggregateID: id,
		occurredAt:  time.Now(),
		Code:        code,
	}
}

func (e ErrorCodeDeleted) EventName() string     { return "errorcode.deleted" }
func (e ErrorCodeDeleted) OccurredAt() time.Time  { return e.occurredAt }
func (e ErrorCodeDeleted) AggregateID() uuid.UUID { return e.aggregateID }
