package command

import (
	"context"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChangeRoleHandler_Handle(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == user.TypedID() {
				return user, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewChangeRoleHandler(repo, eventBus, log)

	newRoleID := uuid.New()
	err := handler.Handle(context.Background(), ChangeRoleCommand{
		UserID: userentity.UserID(user.ID()),
		RoleID: newRoleID,
	})
	require.NoError(t, err)

	if repo.updatedUser == nil {
		t.Fatal("expected user to be updated")
	}

	if repo.updatedUser.RoleID() == nil || *repo.updatedUser.RoleID() != newRoleID {
		t.Error("expected role ID to be updated")
	}

	found := false
	for _, e := range eventBus.publishedEvents {
		if e.EventName() == "user.role_changed" {
			found = true
		}
	}
	if !found {
		t.Error("expected user.role_changed event")
	}
}

func TestChangeRoleHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewChangeRoleHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), ChangeRoleCommand{
		UserID: userentity.NewUserID(),
		RoleID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
