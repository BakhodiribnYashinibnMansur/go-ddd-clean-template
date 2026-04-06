package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

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
