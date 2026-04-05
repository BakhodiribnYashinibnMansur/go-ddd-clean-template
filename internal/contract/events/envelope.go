package events

import (
	"time"

	"github.com/google/uuid"
)

// Envelope is the serialization wrapper every Published Language event carries
// on the bus and in the outbox. It supplies the metadata subscribers need to
// deduplicate, trace, and version-dispatch without inspecting the payload.
//
// Envelope itself is NOT a DomainEvent; domain events embed its fields via
// BaseEvent (see base_event.go) so they remain callable as first-class
// platform/domain.DomainEvent values.
type Envelope struct {
	EventID       uuid.UUID `json:"event_id"`
	EventName     string    `json:"event_name"`
	AggregateID   uuid.UUID `json:"aggregate_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	SchemaVersion int       `json:"schema_version"`
}

// NewEnvelope builds a fresh envelope with a generated EventID.
func NewEnvelope(name string, aggregateID uuid.UUID, version int) Envelope {
	return Envelope{
		EventID:       uuid.New(),
		EventName:     name,
		AggregateID:   aggregateID,
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: version,
	}
}
