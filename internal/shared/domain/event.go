package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents an immutable fact that occurred within the domain.
// Implementations carry enough data for downstream consumers (event handlers, projections, audit logs)
// to react without querying back into the aggregate. Events are collected on AggregateRoot and
// dispatched by the application layer after the transaction commits.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
}
