package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/generic/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoleHandler_Rename(t *testing.T) {
	t.Parallel()

	roleID := domain.NewRoleID()
	existingRole := domain.ReconstructRole(roleID.UUID(), time.Now(), time.Now(), nil, "old_name", nil, nil)

	repo := &mockRoleRepository{
		findByIDFn: func(_ context.Context, id domain.RoleID) (*domain.Role, error) {
			if id == roleID {
				return existingRole, nil
			}
			return nil, domain.ErrRoleNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateRoleHandler(repo, eventBus, log)

	newName := "new_name"
	cmd := UpdateRoleCommand{
		ID:   domain.RoleID(roleID),
		Name: &newName,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedRole == nil {
		t.Fatal("expected role to be updated, but it was nil")
	}

	if repo.updatedRole.Name() != "new_name" {
		t.Errorf("expected name 'new_name', got '%s'", repo.updatedRole.Name())
	}
}

func TestUpdateRoleHandler_SetDescription(t *testing.T) {
	t.Parallel()

	roleID := domain.NewRoleID()
	existingRole := domain.ReconstructRole(roleID.UUID(), time.Now(), time.Now(), nil, "admin", nil, nil)

	repo := &mockRoleRepository{
		findByIDFn: func(_ context.Context, id domain.RoleID) (*domain.Role, error) {
			if id == roleID {
				return existingRole, nil
			}
			return nil, domain.ErrRoleNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateRoleHandler(repo, eventBus, log)

	desc := "Updated description"
	cmd := UpdateRoleCommand{
		ID:          domain.RoleID(roleID),
		Description: &desc,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedRole == nil {
		t.Fatal("expected role to be updated")
	}

	if repo.updatedRole.Description() == nil || *repo.updatedRole.Description() != "Updated description" {
		t.Error("expected description to be 'Updated description'")
	}
}

func TestUpdateRoleHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRoleRepository{} // default findByIDFn returns ErrRoleNotFound
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateRoleHandler(repo, eventBus, log)

	newName := "anything"
	cmd := UpdateRoleCommand{
		ID:   domain.RoleID(uuid.New()),
		Name: &newName,
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got: %v", err)
	}

	if repo.updatedRole != nil {
		t.Error("expected no role to be updated when not found")
	}
}
