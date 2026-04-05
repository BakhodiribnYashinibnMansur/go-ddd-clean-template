package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"gct/internal/platform/domain"
)

func TestNewAggregateRoot_ZeroEvents(t *testing.T) {
	ar := domain.NewAggregateRoot()

	if len(ar.Events()) != 0 {
		t.Errorf("expected 0 events, got %d", len(ar.Events()))
	}
	if ar.ID() == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestAggregateRoot_AddAndClearEvents(t *testing.T) {
	ar := domain.NewAggregateRoot()

	evt := NewTestEvent("order.placed", ar.ID())
	ar.AddEvent(evt)

	if len(ar.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ar.Events()))
	}
	if ar.Events()[0].EventName() != "order.placed" {
		t.Errorf("expected event name 'order.placed', got %q", ar.Events()[0].EventName())
	}

	ar.ClearEvents()
	if len(ar.Events()) != 0 {
		t.Errorf("expected 0 events after clear, got %d", len(ar.Events()))
	}
}

func TestNewAggregateRootWithID(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	ar := domain.NewAggregateRootWithID(id, now, now, nil)

	if ar.ID() != id {
		t.Errorf("expected ID %s, got %s", id, ar.ID())
	}
	if len(ar.Events()) != 0 {
		t.Errorf("expected 0 events, got %d", len(ar.Events()))
	}
}
