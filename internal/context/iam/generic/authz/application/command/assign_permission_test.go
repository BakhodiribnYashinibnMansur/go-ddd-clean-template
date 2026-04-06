package command

import (
	"context"
	"testing"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"

	"github.com/stretchr/testify/require"
)

// --- Mock RolePermissionRepository ---

type mockRolePermissionRepository struct {
	assignedRoleID authzentity.RoleID
	assignedPermID authzentity.PermissionID
	assignFn       func(ctx context.Context, roleID authzentity.RoleID, permissionID authzentity.PermissionID) error
}

func (m *mockRolePermissionRepository) Assign(ctx context.Context, roleID authzentity.RoleID, permissionID authzentity.PermissionID) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, roleID, permissionID)
	}
	m.assignedRoleID = roleID
	m.assignedPermID = permissionID
	return nil
}

func (m *mockRolePermissionRepository) Revoke(ctx context.Context, roleID authzentity.RoleID, permissionID authzentity.PermissionID) error {
	return nil
}

// --- Tests ---

func TestAssignPermissionHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockRolePermissionRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewAssignPermissionHandler(repo, eventBus, log)

	roleID := authzentity.NewRoleID()
	permID := authzentity.NewPermissionID()
	cmd := AssignPermissionCommand{
		RoleID:       roleID,
		PermissionID: permID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.assignedRoleID != roleID {
		t.Errorf("expected role ID %s, got %s", roleID, repo.assignedRoleID)
	}

	if repo.assignedPermID != permID {
		t.Errorf("expected permission ID %s, got %s", permID, repo.assignedPermID)
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "authz.permission_granted" {
		t.Errorf("expected event authz.permission_granted, got %s", eventBus.publishedEvents[0].EventName())
	}

	if eventBus.publishedEvents[0].AggregateID() != roleID.UUID() {
		t.Errorf("expected aggregate ID %s, got %s", roleID, eventBus.publishedEvents[0].AggregateID())
	}
}
