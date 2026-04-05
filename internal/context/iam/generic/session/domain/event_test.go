package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/iam/generic/session/domain"

	"github.com/google/uuid"
)

func TestNewSessionRevokeRequested(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	before := time.Now()

	event := domain.NewSessionRevokeRequested(userID, sessionID)

	if event.EventName() != "session.revoke_requested" {
		t.Fatalf("expected session.revoke_requested, got %s", event.EventName())
	}
	if event.AggregateID() != userID {
		t.Fatalf("expected aggregateID %s, got %s", userID, event.AggregateID())
	}
	if event.SessionID != sessionID {
		t.Fatalf("expected sessionID %s, got %s", sessionID, event.SessionID)
	}
	if event.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestNewSessionRevokeAllRequested(t *testing.T) {
	userID := uuid.New()
	before := time.Now()

	event := domain.NewSessionRevokeAllRequested(userID)

	if event.EventName() != "session.revoke_all_requested" {
		t.Fatalf("expected session.revoke_all_requested, got %s", event.EventName())
	}
	if event.AggregateID() != userID {
		t.Fatalf("expected aggregateID %s, got %s", userID, event.AggregateID())
	}
	if event.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestSessionRevokeRequested_DifferentSessions(t *testing.T) {
	userID := uuid.New()
	s1 := uuid.New()
	s2 := uuid.New()

	e1 := domain.NewSessionRevokeRequested(userID, s1)
	e2 := domain.NewSessionRevokeRequested(userID, s2)

	if e1.SessionID == e2.SessionID {
		t.Fatal("events should have different session IDs")
	}
	if e1.AggregateID() != e2.AggregateID() {
		t.Fatal("events should have same aggregate ID (user)")
	}
}
