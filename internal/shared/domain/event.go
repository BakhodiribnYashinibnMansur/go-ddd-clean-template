package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents something that happened in the domain.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
}
