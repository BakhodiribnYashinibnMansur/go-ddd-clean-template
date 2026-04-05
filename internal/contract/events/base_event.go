package events

import (
	"time"

	"github.com/google/uuid"
)

// BaseEvent embeds the common envelope fields and implements the
// platform/domain.DomainEvent interface so each versioned contract event
// can be published on the shared EventBus without repeating boilerplate.
type BaseEvent struct {
	Envelope
}

// EventName returns the stable event name used for bus routing.
func (b BaseEvent) EventName() string { return b.Envelope.EventName }

// OccurredAt is the UTC timestamp the event was produced.
func (b BaseEvent) OccurredAt() time.Time { return b.Envelope.OccurredAt }

// AggregateID is the identifier of the aggregate root that raised the event.
func (b BaseEvent) AggregateID() uuid.UUID { return b.Envelope.AggregateID }
