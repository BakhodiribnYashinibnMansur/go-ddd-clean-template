package event_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/generic/session/domain/event"

	"github.com/google/uuid"
)

func TestNewSessionRevokeRequested(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	before := time.Now()

	e := event.NewSessionRevokeRequested(userID, sessionID)

	if e.EventName() != "session.revoke_requested" {
		t.Fatalf("expected session.revoke_requested, got %s", e.EventName())
	}
	if e.AggregateID() != userID {
		t.Fatalf("expected aggregateID %s, got %s", userID, e.AggregateID())
	}
	if e.SessionID != sessionID {
		t.Fatalf("expected sessionID %s, got %s", sessionID, e.SessionID)
	}
	if e.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestNewSessionRevokeAllRequested(t *testing.T) {
	userID := uuid.New()
	before := time.Now()

	e := event.NewSessionRevokeAllRequested(userID)

	if e.EventName() != "session.revoke_all_requested" {
		t.Fatalf("expected session.revoke_all_requested, got %s", e.EventName())
	}
	if e.AggregateID() != userID {
		t.Fatalf("expected aggregateID %s, got %s", userID, e.AggregateID())
	}
	if e.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestSessionRevokeRequested_DifferentSessions(t *testing.T) {
	userID := uuid.New()
	s1 := uuid.New()
	s2 := uuid.New()

	e1 := event.NewSessionRevokeRequested(userID, s1)
	e2 := event.NewSessionRevokeRequested(userID, s2)

	if e1.SessionID == e2.SessionID {
		t.Fatal("events should have different session IDs")
	}
	if e1.AggregateID() != e2.AggregateID() {
		t.Fatal("events should have same aggregate ID (user)")
	}
}
