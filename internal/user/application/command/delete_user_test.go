package command

import (
	"context"
	"testing"

	"gct/internal/user/domain"

	"github.com/google/uuid"
)

func TestDeleteUserHandler_Handle(t *testing.T) {
	user := makeTestUser(t)
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

	handler := NewDeleteUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), DeleteUserCommand{ID: user.ID()})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteUserHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), DeleteUserCommand{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
