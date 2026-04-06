package entity

import (
	"testing"

	"gct/internal/context/iam/generic/authz/domain/event"

	"github.com/google/uuid"
)

func TestContract_Role_AddPermission_RaisesEvent(t *testing.T) {
	role := NewRole("admin")
	role.ClearEvents() // clear the RoleCreated event

	perm := NewPermission("users.read", nil)
	role.AddPermission(*perm)

	// Verify the permission was added.
	perms := role.Permissions()
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(perms))
	}
	if perms[0].Name() != "users.read" {
		t.Errorf("expected permission name %q, got %q", "users.read", perms[0].Name())
	}

	// Verify PermissionGranted event was raised.
	events := role.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	evt, ok := events[0].(event.PermissionGranted)
	if !ok {
		t.Fatalf("expected PermissionGranted event, got %T", events[0])
	}
	if evt.EventName() != "authz.permission_granted" {
		t.Errorf("expected event name %q, got %q", "authz.permission_granted", evt.EventName())
	}
	if evt.PermissionID != perm.ID() {
		t.Errorf("expected permission ID %v in event, got %v", perm.ID(), evt.PermissionID)
	}
}

func TestContract_Role_RemovePermission_NotFound(t *testing.T) {
	role := NewRole("viewer")

	err := role.RemovePermission(uuid.New())
	if err == nil {
		t.Fatal("expected error when removing non-existent permission")
	}
	if err != ErrPermissionNotFound {
		t.Errorf("expected ErrPermissionNotFound, got %v", err)
	}
}

func TestContract_Permission_AddScope(t *testing.T) {
	perm := NewPermission("users.manage", nil)

	scope := Scope{Path: "/api/v1/users", Method: "GET"}
	perm.AddScope(scope)

	scopes := perm.Scopes()
	if len(scopes) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(scopes))
	}
	if scopes[0].Path != "/api/v1/users" {
		t.Errorf("expected scope path %q, got %q", "/api/v1/users", scopes[0].Path)
	}
	if scopes[0].Method != "GET" {
		t.Errorf("expected scope method %q, got %q", "GET", scopes[0].Method)
	}
}

func TestContract_Permission_RemoveScope_NotFound(t *testing.T) {
	perm := NewPermission("users.manage", nil)

	err := perm.RemoveScope("/api/v1/nonexistent", "DELETE")
	if err == nil {
		t.Fatal("expected error when removing non-existent scope")
	}
	if err != ErrScopeNotFound {
		t.Errorf("expected ErrScopeNotFound, got %v", err)
	}
}

func TestContract_Policy_Toggle(t *testing.T) {
	permID := NewPermissionID()
	policy := NewPolicy(permID.UUID(), PolicyAllow)

	// New policy should be active by default.
	if !policy.IsActive() {
		t.Fatal("expected new policy to be active")
	}

	// First toggle: active -> inactive.
	policy.Toggle()
	if policy.IsActive() {
		t.Fatal("expected policy to be inactive after first Toggle")
	}

	// Second toggle: inactive -> active.
	policy.Toggle()
	if !policy.IsActive() {
		t.Fatal("expected policy to be active after second Toggle")
	}
}

func TestContract_Policy_Effects(t *testing.T) {
	permID := NewPermissionID()

	allowPolicy := NewPolicy(permID.UUID(), PolicyAllow)
	if allowPolicy.Effect() != PolicyAllow {
		t.Errorf("expected effect %q, got %q", PolicyAllow, allowPolicy.Effect())
	}

	denyPolicy := NewPolicy(permID.UUID(), PolicyDeny)
	if denyPolicy.Effect() != PolicyDeny {
		t.Errorf("expected effect %q, got %q", PolicyDeny, denyPolicy.Effect())
	}

	// Verify the string representations.
	if string(PolicyAllow) != "ALLOW" {
		t.Errorf("expected PolicyAllow to be %q, got %q", "ALLOW", string(PolicyAllow))
	}
	if string(PolicyDeny) != "DENY" {
		t.Errorf("expected PolicyDeny to be %q, got %q", "DENY", string(PolicyDeny))
	}
}

func TestContract_Scope_Immutable(t *testing.T) {
	scope := Scope{Path: "/api/v1/roles", Method: "POST"}

	// Scope is a value object (struct with exported fields).
	// Verify that copying a scope produces an independent value.
	copied := scope
	copied.Path = "/api/v1/changed"
	copied.Method = "DELETE"

	if scope.Path != "/api/v1/roles" {
		t.Errorf("original scope path mutated: expected %q, got %q", "/api/v1/roles", scope.Path)
	}
	if scope.Method != "POST" {
		t.Errorf("original scope method mutated: expected %q, got %q", "POST", scope.Method)
	}

	// Verify that a Scope is a plain struct value object: assigning it
	// produces an independent copy whose fields do not alias the original.
	perm := NewPermission("test.perm", nil)
	perm.AddScope(scope)

	scopes := perm.Scopes()
	snapshot := scopes[0] // copy the value
	snapshot.Path = "/api/v1/tampered"
	snapshot.Method = "DELETE"

	// The copy must not have affected the original stored in the permission.
	if perm.Scopes()[0].Path != "/api/v1/roles" {
		t.Error("scope value inside permission was mutated through a copy; value semantics violated")
	}
	if perm.Scopes()[0].Method != "POST" {
		t.Error("scope method inside permission was mutated through a copy; value semantics violated")
	}
}
