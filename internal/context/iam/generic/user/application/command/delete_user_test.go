package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/user/domain"

	"github.com/stretchr/testify/require"
)

func TestDeleteUserHandler_Handle(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id domain.UserID) (*domain.User, error) {
			if id == user.TypedID() {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), DeleteUserCommand{ID: domain.UserID(user.ID())})
	require.NoError(t, err)

	if repo.updatedUser == nil {
		t.Fatal("expected user to be updated (soft-deleted)")
	}

	if repo.updatedUser.IsActive() {
		t.Error("expected user to be inactive after delete")
	}

	if repo.updatedUser.DeletedAt() == nil {
		t.Error("expected deletedAt to be set after soft-delete")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Error("expected events to be published")
	}
}

func TestDeleteUserHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), DeleteUserCommand{ID: domain.NewUserID()})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
