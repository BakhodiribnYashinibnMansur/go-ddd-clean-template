package domain

import (
	"time"

	"github.com/google/uuid"
)

// AggregateRoot is the base for all aggregate roots in the domain.
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

// ClearEvents removes all recorded domain events.
func (a *AggregateRoot) ClearEvents() {
	a.events = make([]DomainEvent, 0)
}
