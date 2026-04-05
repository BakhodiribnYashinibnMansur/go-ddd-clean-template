package eventbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
)

// testEvent implements domain.DomainEvent for testing.
type testEvent struct {
	name        string
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func (e testEvent) EventName() string       { return e.name }
func (e testEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e testEvent) AggregateID() uuid.UUID  { return e.aggregateID }

func newTestEvent(name string) testEvent {
	return testEvent{
		name:        name,
		aggregateID: uuid.New(),
		occurredAt:  time.Now(),
	}
}

func TestPublish_HandlerReceivesEvent(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus()

	var received domain.DomainEvent
	err := bus.Subscribe("user.created", func(ctx context.Context, event domain.DomainEvent) error {
		received = event
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	evt := newTestEvent("user.created")
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if received == nil {
		t.Fatal("handler was not called")
	}
	if received.EventName() != "user.created" {
		t.Errorf("expected event name 'user.created', got %q", received.EventName())
	}
	if received.AggregateID() != evt.AggregateID() {
		t.Errorf("aggregate ID mismatch")
	}
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus()

	callCount := 0
	handler := func(ctx context.Context, event domain.DomainEvent) error {
		callCount++
		return nil
	}

	_ = bus.Subscribe("order.placed", handler)
	_ = bus.Subscribe("order.placed", handler)

	if err := bus.Publish(context.Background(), newTestEvent("order.placed")); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 handler calls, got %d", callCount)
	}
}

func TestPublish_NoSubscribers_NoError(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus()

	err := bus.Publish(context.Background(), newTestEvent("nobody.listens"))
	if err != nil {
		t.Fatalf("expected no error for unsubscribed event, got: %v", err)
	}
}

func TestPublish_HandlerError_FailsFast(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus()

	expectedErr := errors.New("handler failed")
	_ = bus.Subscribe("fail.event", func(ctx context.Context, event domain.DomainEvent) error {
		return expectedErr
	})

	secondCalled := false
	_ = bus.Subscribe("fail.event", func(ctx context.Context, event domain.DomainEvent) error {
		secondCalled = true
		return nil
	})

	err := bus.Publish(context.Background(), newTestEvent("fail.event"))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got: %v", err)
	}
	if secondCalled {
		t.Error("second handler should not have been called after first failed")
	}
}

func TestSubscribe_MultipleEvents(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus()

	events := make(map[string]bool)
	handler := func(ctx context.Context, event domain.DomainEvent) error {
		events[event.EventName()] = true
		return nil
	}

	_ = bus.Subscribe("event.a", handler)
	_ = bus.Subscribe("event.b", handler)

	_ = bus.Publish(context.Background(), newTestEvent("event.a"))
	_ = bus.Publish(context.Background(), newTestEvent("event.b"))

	if !events["event.a"] || !events["event.b"] {
		t.Errorf("expected both events to be handled, got: %v", events)
	}
}
