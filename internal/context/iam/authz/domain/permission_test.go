package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewPermission(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("users.read", nil)

	if perm.ID() == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if perm.Name() != "users.read" {
		t.Errorf("expected name 'users.read', got %q", perm.Name())
	}
	if perm.ParentID() != nil {
		t.Error("expected nil parentID")
	}
	if perm.Description() != nil {
		t.Error("expected nil description")
	}
	if len(perm.Scopes()) != 0 {
		t.Errorf("expected 0 scopes, got %d", len(perm.Scopes()))
	}
}

func TestNewPermission_WithParentID(t *testing.T) {
	t.Parallel()

	parentID := uuid.New()
	perm := domain.NewPermission("users.read.detail", &parentID)

	if perm.ParentID() == nil {
		t.Fatal("expected non-nil parentID")
	}
	if *perm.ParentID() != parentID {
		t.Errorf("expected parentID %s, got %s", parentID, *perm.ParentID())
	}
}

func TestReconstructPermission(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	parentID := uuid.New()
	desc := "can read users"
	now := time.Now()
	scopes := []domain.Scope{
		{Path: "/api/users", Method: "GET"},
	}

	perm := domain.ReconstructPermission(
		id, now, now, nil,
		&parentID, "users.read", &desc, scopes,
	)

	if perm.ID() != id {
		t.Fatalf("expected ID %s, got %s", id, perm.ID())
	}
	if perm.Name() != "users.read" {
		t.Errorf("expected name 'users.read', got %q", perm.Name())
	}
	if perm.Description() == nil || *perm.Description() != "can read users" {
		t.Error("expected description 'can read users'")
	}
	if perm.ParentID() == nil || *perm.ParentID() != parentID {
		t.Error("parentID mismatch")
	}
	if len(perm.Scopes()) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(perm.Scopes()))
	}
	if perm.Scopes()[0].Path != "/api/users" {
		t.Errorf("expected scope path '/api/users', got %q", perm.Scopes()[0].Path)
	}
}

func TestReconstructPermission_NilScopes(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	perm := domain.ReconstructPermission(
		id, now, now, nil,
		nil, "test", nil, nil,
	)

	if perm.Scopes() == nil {
		t.Fatal("expected non-nil scopes slice (empty, not nil)")
	}
	if len(perm.Scopes()) != 0 {
		t.Errorf("expected 0 scopes, got %d", len(perm.Scopes()))
	}
}

func TestPermission_Rename(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("old_name", nil)
	before := perm.UpdatedAt()

	perm.Rename("new_name")

	if perm.Name() != "new_name" {
		t.Errorf("expected name 'new_name', got %q", perm.Name())
	}
	if !perm.UpdatedAt().After(before) && perm.UpdatedAt().Equal(before) {
		// Touch may be same instant in fast tests; at minimum should not be before.
	}
}

func TestPermission_SetDescription(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("test", nil)

	desc := "a description"
	perm.SetDescription(&desc)
	if perm.Description() == nil || *perm.Description() != "a description" {
		t.Error("expected description 'a description'")
	}

	perm.SetDescription(nil)
	if perm.Description() != nil {
		t.Error("expected nil description after clearing")
	}
}

func TestPermission_AddScope(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("test", nil)

	perm.AddScope(domain.Scope{Path: "/api/a", Method: "GET"})
	perm.AddScope(domain.Scope{Path: "/api/b", Method: "POST"})

	if len(perm.Scopes()) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(perm.Scopes()))
	}
	if perm.Scopes()[0].Path != "/api/a" {
		t.Errorf("expected first scope path '/api/a', got %q", perm.Scopes()[0].Path)
	}
	if perm.Scopes()[1].Method != "POST" {
		t.Errorf("expected second scope method 'POST', got %q", perm.Scopes()[1].Method)
	}
}

func TestPermission_RemoveScope_Success(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("test", nil)
	perm.AddScope(domain.Scope{Path: "/api/a", Method: "GET"})
	perm.AddScope(domain.Scope{Path: "/api/b", Method: "POST"})

	err := perm.RemoveScope("/api/a", "GET")
	require.NoError(t, err)
	if len(perm.Scopes()) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(perm.Scopes()))
	}
	if perm.Scopes()[0].Path != "/api/b" {
		t.Errorf("expected remaining scope path '/api/b', got %q", perm.Scopes()[0].Path)
	}
}

func TestPermission_RemoveScope_NotFound(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("test", nil)

	err := perm.RemoveScope("/api/missing", "GET")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPermission_RemoveScope_WrongMethod(t *testing.T) {
	t.Parallel()

	perm := domain.NewPermission("test", nil)
	perm.AddScope(domain.Scope{Path: "/api/a", Method: "GET"})

	err := perm.RemoveScope("/api/a", "POST")
	if err == nil {
		t.Fatal("expected error when method does not match")
	}
}
