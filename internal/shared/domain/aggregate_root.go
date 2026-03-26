package domain

import (
	"time"

	"github.com/google/uuid"
)

// AggregateRoot extends BaseEntity with domain event collection, forming the transactional consistency boundary.
// Events accumulate during a business operation and should be dispatched by the application layer after persistence.
// Callers must invoke ClearEvents after dispatching to prevent duplicate processing.
type AggregateRoot struct {
	BaseEntity
	events []DomainEvent
}

// NewAggregateRoot creates a new AggregateRoot with a generated UUID and empty events.
func NewAggregateRoot() AggregateRoot {
	return AggregateRoot{
		BaseEntity: NewBaseEntity(),
		events:     make([]DomainEvent, 0),
	}
}

// NewAggregateRootWithID reconstructs an AggregateRoot from persisted data.
func NewAggregateRootWithID(id uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) AggregateRoot {
	return AggregateRoot{
		BaseEntity: NewBaseEntityWithID(id, createdAt, updatedAt, deletedAt),
		events:     make([]DomainEvent, 0),
	}
}

// AddEvent records a domain event.
func (a *AggregateRoot) AddEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// Events returns all recorded domain events.
func (a *AggregateRoot) Events() []DomainEvent {
	return a.events
}

// ClearEvents removes all recorded domain events. Must be called after events are dispatched
// to prevent double-publishing on subsequent Save/Update calls.
func (a *AggregateRoot) ClearEvents() {
	a.events = make([]DomainEvent, 0)
}
