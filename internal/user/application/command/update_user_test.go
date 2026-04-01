package command

import (
	"context"
	"testing"
	"time"

	"gct/internal/user/domain"

	"github.com/google/uuid"
)

func makeTestUser(t *testing.T) *domain.User {
	t.Helper()
	phone, err := domain.NewPhone("+998901234567")
	if err != nil {
		t.Fatalf("NewPhone: %v", err)
	}
	pw, err := domain.NewPasswordFromRaw("StrongP@ss123")
	if err != nil {
		t.Fatalf("NewPasswordFromRaw: %v", err)
	}
	email, _ := domain.NewEmail("old@example.com")
	username := "olduser"
	return domain.ReconstructUser(
		uuid.New(),
		time.Now(), time.Now(), nil,
		phone, &email, &username, pw,
		nil, map[string]string{"key": "val"},
		true, false, nil, nil,
	)
}

func TestUpdateUserHandler_Handle(t *testing.T) {
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

	handler := NewUpdateUserHandler(repo, eventBus, log)

	newEmail := "new@example.com"
	newUsername := "newuser"
	cmd := UpdateUserCommand{
		ID:       user.ID(),
		Email:    &newEmail,
		Username: &newUsername,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updatedUser == nil {
		t.Fatal("expected user to be updated")
	}

	if repo.updatedUser.Email() == nil || repo.updatedUser.Email().Value() != "new@example.com" {
		t.Error("expected email to be updated")
	}

	if repo.updatedUser.Username() == nil || *repo.updatedUser.Username() != "newuser" {
		t.Error("expected username to be updated")
	}
}

func TestUpdateUserHandler_NotFound(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	cmd := UpdateUserCommand{ID: uuid.New()}
	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestUpdateUserHandler_InvalidEmail(t *testing.T) {
	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	badEmail := "not-an-email"
	cmd := UpdateUserCommand{
		ID:    user.ID(),
		Email: &badEmail,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestUpdateUserHandler_OnlyAttributes(t *testing.T) {
	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	cmd := UpdateUserCommand{
		ID:         user.ID(),
		Attributes: map[string]string{"new_key": "new_val"},
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updatedUser.Attributes()["new_key"] != "new_val" {
		t.Error("expected attributes to be updated")
	}
}
