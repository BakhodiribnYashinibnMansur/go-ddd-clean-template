package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestApproveUserHandler_Handle(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t) // not approved by default
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == user.ID() {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewApproveUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), ApproveUserCommand{ID: domain.UserID(user.ID())})
	require.NoError(t, err)

	if repo.updatedUser == nil {
		t.Fatal("expected user to be updated")
	}

	if !repo.updatedUser.IsApproved() {
		t.Error("expected user to be approved")
	}

	found := false
	for _, e := range eventBus.publishedEvents {
		if e.EventName() == "user.approved" {
			found = true
		}
	}
	if !found {
		t.Error("expected user.approved event")
	}
}

func TestApproveUserHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewApproveUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), ApproveUserCommand{ID: domain.NewUserID()})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
