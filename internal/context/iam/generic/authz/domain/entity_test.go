package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewRole(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")

	if role.Name() != "admin" {
		t.Errorf("expected name 'admin', got '%s'", role.Name())
	}

	if role.ID() == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if len(role.Permissions()) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(role.Permissions()))
	}

	if len(role.Events()) == 0 {
		t.Fatal("expected at least one event")
	}

	if role.Events()[0].EventName() != "authz.role_created" {
		t.Errorf("expected event authz.role_created, got %s", role.Events()[0].EventName())
	}
}

func TestRole_Rename(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")
	role.Rename("super_admin")

	if role.Name() != "super_admin" {
		t.Errorf("expected name 'super_admin', got '%s'", role.Name())
	}
}

func TestRole_SetDescription(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")
	desc := "Administrator role"
	role.SetDescription(&desc)

	if role.Description() == nil || *role.Description() != "Administrator role" {
		t.Error("expected description to be 'Administrator role'")
	}
}

func TestRole_AddPermission(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")
	perm := NewPermission("users.read", nil)

	role.AddPermission(*perm)

	if len(role.Permissions()) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(role.Permissions()))
	}

	if role.Permissions()[0].Name() != "users.read" {
		t.Errorf("expected permission name 'users.read', got '%s'", role.Permissions()[0].Name())
	}

	// Should have RoleCreated + PermissionGranted events.
	if len(role.Events()) != 2 {
		t.Errorf("expected 2 events, got %d", len(role.Events()))
	}
}

func TestRole_RemovePermission(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")
	perm := NewPermission("users.read", nil)
	role.AddPermission(*perm)

	err := role.RemovePermission(perm.ID())
	require.NoError(t, err)

	if len(role.Permissions()) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(role.Permissions()))
	}
}

func TestRole_RemovePermission_NotFound(t *testing.T) {
	t.Parallel()

	role := NewRole("admin")

	err := role.RemovePermission(uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrPermissionNotFound {
		t.Errorf("expected ErrPermissionNotFound, got %v", err)
	}
}

func TestPermission_AddScope(t *testing.T) {
	t.Parallel()

	perm := NewPermission("users.read", nil)
	scope := Scope{Path: "/api/users", Method: "GET"}
	perm.AddScope(scope)

	if len(perm.Scopes()) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(perm.Scopes()))
	}

	if perm.Scopes()[0].Path != "/api/users" {
		t.Errorf("expected path '/api/users', got '%s'", perm.Scopes()[0].Path)
	}
}

func TestPermission_RemoveScope(t *testing.T) {
	t.Parallel()

	perm := NewPermission("users.read", nil)
	perm.AddScope(Scope{Path: "/api/users", Method: "GET"})

	err := perm.RemoveScope("/api/users", "GET")
	require.NoError(t, err)

	if len(perm.Scopes()) != 0 {
		t.Errorf("expected 0 scopes, got %d", len(perm.Scopes()))
	}
}

func TestPermission_RemoveScope_NotFound(t *testing.T) {
	t.Parallel()

	perm := NewPermission("users.read", nil)

	err := perm.RemoveScope("/api/missing", "GET")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrScopeNotFound {
		t.Errorf("expected ErrScopeNotFound, got %v", err)
	}
}
