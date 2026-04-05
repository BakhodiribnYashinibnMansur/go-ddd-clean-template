package domain

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

func TestReconstructRole_Full(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	desc := "Admin"
	perms := []Permission{*NewPermission("test", nil)}

	role := ReconstructRole(id, now, now, nil, "admin", &desc, perms)

	if role.ID() != id {
		t.Errorf("expected ID %s, got %s", id, role.ID())
	}
	if role.Name() != "admin" {
		t.Errorf("expected name admin, got %s", role.Name())
	}
	if role.Description() == nil || *role.Description() != "Admin" {
		t.Error("expected description 'Admin'")
	}
	if len(role.Permissions()) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(role.Permissions()))
	}
	if len(role.Events()) != 0 {
		t.Errorf("expected 0 events on reconstruct, got %d", len(role.Events()))
	}
}

func TestReconstructRole_NilPermissions(t *testing.T) {
	t.Parallel()

	role := ReconstructRole(uuid.New(), time.Now(), time.Now(), nil, "test", nil, nil)
	if role.Permissions() == nil {
		t.Error("expected non-nil permissions")
	}
}

func TestReconstructPolicy_Full(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	permID := uuid.New()
	now := time.Now()
	conds := map[string]any{"ip": "10.0.0.0/8"}

	policy := ReconstructPolicy(id, now, now, nil, permID, PolicyDeny, 50, false, conds)

	if policy.ID() != id {
		t.Errorf("expected ID %s, got %s", id, policy.ID())
	}
	if policy.PermissionID() != permID {
		t.Errorf("expected permissionID %s, got %s", permID, policy.PermissionID())
	}
	if policy.Effect() != PolicyDeny {
		t.Errorf("expected DENY, got %s", policy.Effect())
	}
	if policy.Priority() != 50 {
		t.Errorf("expected priority 50, got %d", policy.Priority())
	}
	if policy.IsActive() {
		t.Error("expected inactive")
	}
	if policy.Conditions()["ip"] != "10.0.0.0/8" {
		t.Error("expected ip condition")
	}
}

func TestReconstructPolicy_NilConditions(t *testing.T) {
	t.Parallel()

	policy := ReconstructPolicy(uuid.New(), time.Now(), time.Now(), nil, uuid.New(), PolicyAllow, 0, true, nil)
	if policy.Conditions() == nil {
		t.Error("expected non-nil conditions")
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
