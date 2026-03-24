package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"gct/internal/shared/application"
	"gct/internal/shared/domain"
)

// testEvent is a fake domain event for testing.
type testEvent struct {
	name        string
	occurredAt  time.Time
	aggregateID uuid.UUID
}

func (e testEvent) EventName() string       { return e.name }
func (e testEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e testEvent) AggregateID() uuid.UUID  { return e.aggregateID }

// mockEventBus records published events and subscribed handlers.
type mockEventBus struct {
	published  []domain.DomainEvent
	subscribed map[string]application.EventHandler
}

func newMockEventBus() *mockEventBus {
	return &mockEventBus{
		subscribed: make(map[string]application.EventHandler),
	}
}

func (m *mockEventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}

func (m *mockEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	m.subscribed[eventName] = handler
	return nil
}

// Compile-time interface satisfaction check.
var _ application.EventBus = (*mockEventBus)(nil)

func TestEventBus_Publish(t *testing.T) {
	bus := newMockEventBus()

	evt1 := testEvent{name: "UserCreated", occurredAt: time.Now(), aggregateID: uuid.New()}
	evt2 := testEvent{name: "UserUpdated", occurredAt: time.Now(), aggregateID: uuid.New()}

	err := bus.Publish(context.Background(), evt1, evt2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(bus.published) != 2 {
		t.Fatalf("expected 2 published events, got %d", len(bus.published))
	}
	if bus.published[0].EventName() != "UserCreated" {
		t.Fatalf("expected first event UserCreated, got %s", bus.published[0].EventName())
	}
	if bus.published[1].EventName() != "UserUpdated" {
		t.Fatalf("expected second event UserUpdated, got %s", bus.published[1].EventName())
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	bus := newMockEventBus()

	handler := func(ctx context.Context, event domain.DomainEvent) error {
		return nil
	}

	err := bus.Subscribe("UserCreated", handler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := bus.subscribed["UserCreated"]; !ok {
		t.Fatal("expected handler to be subscribed for UserCreated")
	}
}
