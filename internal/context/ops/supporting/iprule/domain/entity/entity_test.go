package entity_test

import (
	"testing"
	"time"

	"gct/internal/context/ops/supporting/iprule/domain/entity"

	"github.com/google/uuid"
)

func TestNewIPRule(t *testing.T) {
	t.Parallel()

	r := entity.NewIPRule("192.168.1.1", "DENY", "suspicious activity", nil)

	if r.IPAddress() != "192.168.1.1" {
		t.Fatalf("expected ip 192.168.1.1, got %s", r.IPAddress())
	}
	if r.Action() != "DENY" {
		t.Fatalf("expected action DENY, got %s", r.Action())
	}
	if r.Reason() != "suspicious activity" {
		t.Fatalf("expected reason suspicious activity, got %s", r.Reason())
	}
	if r.ExpiresAt() != nil {
		t.Fatal("expiresAt should be nil")
	}

	events := r.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "iprule.created" {
		t.Fatalf("expected iprule.created, got %s", events[0].EventName())
	}
}

func TestIPRule_Update(t *testing.T) {
	t.Parallel()

	r := entity.NewIPRule("10.0.0.1", "ALLOW", "trusted", nil)
	r.ClearEvents()

	newAction := "DENY"
	r.Update(nil, &newAction, nil, nil)

	if r.Action() != "DENY" {
		t.Fatalf("expected action DENY, got %s", r.Action())
	}
}

func TestReconstructIPRule(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	expires := now.Add(24 * time.Hour)

	r := entity.ReconstructIPRule(id, now, now, "10.0.0.1", "ALLOW", "test", &expires)

	if r.ID() != id {
		t.Fatal("ID mismatch")
	}
	if r.IPAddress() != "10.0.0.1" {
		t.Fatal("ip mismatch")
	}
	if r.ExpiresAt() == nil {
		t.Fatal("expiresAt should not be nil")
	}
	if len(r.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(r.Events()))
	}
}
