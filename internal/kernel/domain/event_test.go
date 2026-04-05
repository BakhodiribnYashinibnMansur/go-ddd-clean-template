package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"gct/internal/kernel/domain"
)

// TestEvent is a concrete DomainEvent for testing.
type TestEvent struct {
	name        string
	occurredAt  time.Time
	aggregateID uuid.UUID
}

func NewTestEvent(name string, aggregateID uuid.UUID) *TestEvent {
	return &TestEvent{
		name:        name,
		occurredAt:  time.Now(),
		aggregateID: aggregateID,
	}
}

func (e *TestEvent) EventName() string      { return e.name }
func (e *TestEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e *TestEvent) AggregateID() uuid.UUID  { return e.aggregateID }

func TestDomainEvent_Interface(t *testing.T) {
	id := uuid.New()
	evt := NewTestEvent("user.created", id)

	// Compile-time check
	var _ domain.DomainEvent = evt

	if evt.EventName() != "user.created" {
		t.Errorf("expected event name 'user.created', got %q", evt.EventName())
	}
	if evt.AggregateID() != id {
		t.Errorf("expected aggregate ID %s, got %s", id, evt.AggregateID())
	}
	if evt.OccurredAt().IsZero() {
		t.Error("expected non-zero occurred time")
	}
}
