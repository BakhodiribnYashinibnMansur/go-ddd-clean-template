package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock PermissionScopeRepository ---

type mockPermissionScopeRepository struct {
	assignedPermID uuid.UUID
	assignedPath   string
	assignedMethod string
	assignFn       func(ctx context.Context, permissionID uuid.UUID, path, method string) error
}

func (m *mockPermissionScopeRepository) Assign(ctx context.Context, permissionID uuid.UUID, path, method string) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, permissionID, path, method)
	}
	m.assignedPermID = permissionID
	m.assignedPath = path
	m.assignedMethod = method
	return nil
}

func (m *mockPermissionScopeRepository) Revoke(ctx context.Context, permissionID uuid.UUID, path, method string) error {
	return nil
}

// --- Tests ---

func TestAssignScopeHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionScopeRepository{}
	log := &mockLogger{}

	handler := NewAssignScopeHandler(repo, log)

	permID := uuid.New()
	cmd := AssignScopeCommand{
		PermissionID: domain.PermissionID(permID),
		Path:         "/api/v1/orders",
		Method:       "POST",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.assignedPermID != permID {
		t.Errorf("expected permission ID %s, got %s", permID, repo.assignedPermID)
	}

	if repo.assignedPath != "/api/v1/orders" {
		t.Errorf("expected path '/api/v1/orders', got '%s'", repo.assignedPath)
	}

	if repo.assignedMethod != "POST" {
		t.Errorf("expected method 'POST', got '%s'", repo.assignedMethod)
	}
}
