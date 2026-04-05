package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
)

func TestUserCreated(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	e := domain.NewUserCreated(id, "+998901234567")
	if e.EventName() != "user.created" {
		t.Fatalf("expected user.created, got %s", e.EventName())
	}
	if e.AggregateID() != id {
		t.Fatal("aggregate ID mismatch")
	}
	if e.OccurredAt().IsZero() {
		t.Fatal("occurredAt should not be zero")
	}
}

func TestUserSignedIn(t *testing.T) {
	t.Parallel()

	uid := domain.NewUserID()
	sid := domain.NewSessionID()
	e := domain.NewUserSignedIn(uid.UUID(), sid.UUID(), "1.2.3.4")
	if e.EventName() != "user.signed_in" {
		t.Fatalf("expected user.signed_in, got %s", e.EventName())
	}
	if e.AggregateID() != uid.UUID() {
		t.Fatal("aggregate ID mismatch")
	}
	if e.SessionID != sid.UUID() {
		t.Fatal("session ID mismatch")
	}
}

func TestUserDeactivated(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	e := domain.NewUserDeactivated(id)
	if e.EventName() != "user.deactivated" {
		t.Fatalf("expected user.deactivated, got %s", e.EventName())
	}
	if e.AggregateID() != id {
		t.Fatal("aggregate ID mismatch")
	}
}

func TestPasswordChanged(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	e := domain.NewPasswordChanged(id)
	if e.EventName() != "user.password_changed" {
		t.Fatalf("expected user.password_changed, got %s", e.EventName())
	}
}

func TestUserApproved(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	e := domain.NewUserApproved(id)
	if e.EventName() != "user.approved" {
		t.Fatalf("expected user.approved, got %s", e.EventName())
	}
}

func TestRoleChanged(t *testing.T) {
	t.Parallel()

	uid := domain.NewUserID()
	oldRole := uuid.New()
	newRole := uuid.New()
	e := domain.NewRoleChanged(uid.UUID(), &oldRole, newRole)
	if e.EventName() != "user.role_changed" {
		t.Fatalf("expected user.role_changed, got %s", e.EventName())
	}
	if *e.OldRoleID != oldRole {
		t.Fatal("old role ID mismatch")
	}
	if e.NewRoleID != newRole {
		t.Fatal("new role ID mismatch")
	}
}

func TestEventsImplementDomainEvent(t *testing.T) {
	t.Parallel()

	// Compile-time check that all events satisfy shared.DomainEvent.
	// The interface requires EventName(), OccurredAt(), AggregateID().
	id := uuid.New()
	events := []interface {
		EventName() string
		OccurredAt() time.Time
		AggregateID() uuid.UUID
	}{
		domain.NewUserCreated(id, "+1234567890"),
		domain.NewUserSignedIn(id, uuid.New(), "1.1.1.1"),
		domain.NewUserDeactivated(id),
		domain.NewPasswordChanged(id),
		domain.NewUserApproved(id),
		domain.NewRoleChanged(id, nil, uuid.New()),
	}
	for _, e := range events {
		if e.EventName() == "" {
			t.Fatal("event name should not be empty")
		}
	}
}
