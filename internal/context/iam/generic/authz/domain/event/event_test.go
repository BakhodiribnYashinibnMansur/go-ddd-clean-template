package event

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRoleCreated(t *testing.T) {
	t.Parallel()

	roleID := uuid.New()
	before := time.Now()
	e := NewRoleCreated(roleID, "admin")

	if e.EventName() != "authz.role_created" {
		t.Errorf("expected event name 'authz.role_created', got %q", e.EventName())
	}
	if e.AggregateID() != roleID {
		t.Errorf("expected aggregate ID %s, got %s", roleID, e.AggregateID())
	}
	if e.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", e.Name)
	}
	if e.OccurredAt().Before(before) {
		t.Error("expected occurredAt to be at or after test start")
	}
}

func TestRoleDeleted(t *testing.T) {
	t.Parallel()

	roleID := uuid.New()
	before := time.Now()
	e := NewRoleDeleted(roleID)

	if e.EventName() != "authz.role_deleted" {
		t.Errorf("expected event name 'authz.role_deleted', got %q", e.EventName())
	}
	if e.AggregateID() != roleID {
		t.Errorf("expected aggregate ID %s, got %s", roleID, e.AggregateID())
	}
	if e.OccurredAt().Before(before) {
		t.Error("expected occurredAt to be at or after test start")
	}
}

func TestPolicyUpdated(t *testing.T) {
	t.Parallel()

	roleID := uuid.New()
	policyID := uuid.New()
	before := time.Now()
	e := NewPolicyUpdated(roleID, policyID)

	if e.EventName() != "authz.policy_updated" {
		t.Errorf("expected event name 'authz.policy_updated', got %q", e.EventName())
	}
	if e.AggregateID() != roleID {
		t.Errorf("expected aggregate ID %s, got %s", roleID, e.AggregateID())
	}
	if e.PolicyID != policyID {
		t.Errorf("expected policy ID %s, got %s", policyID, e.PolicyID)
	}
	if e.OccurredAt().Before(before) {
		t.Error("expected occurredAt to be at or after test start")
	}
}

func TestPermissionGranted(t *testing.T) {
	t.Parallel()

	roleID := uuid.New()
	permID := uuid.New()
	before := time.Now()
	e := NewPermissionGranted(roleID, permID)

	if e.EventName() != "authz.permission_granted" {
		t.Errorf("expected event name 'authz.permission_granted', got %q", e.EventName())
	}
	if e.AggregateID() != roleID {
		t.Errorf("expected aggregate ID %s, got %s", roleID, e.AggregateID())
	}
	if e.PermissionID != permID {
		t.Errorf("expected permission ID %s, got %s", permID, e.PermissionID)
	}
	if e.OccurredAt().Before(before) {
		t.Error("expected occurredAt to be at or after test start")
	}
}
