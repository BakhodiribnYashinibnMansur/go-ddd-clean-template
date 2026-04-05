package command

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func makeTestUser(t *testing.T) *domain.User {
	t.Helper()
	phone, err := domain.NewPhone("+998901234567")
	require.NoError(t, err)
	pw, err := domain.NewPasswordFromRaw("StrongP@ss123")
	require.NoError(t, err)
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

	handler := NewUpdateUserHandler(repo, eventBus, log)

	newEmail := "new@example.com"
	newUsername := "newuser"
	cmd := UpdateUserCommand{
		ID:       domain.UserID(user.ID()),
		Email:    &newEmail,
		Username: &newUsername,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

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
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	cmd := UpdateUserCommand{ID: domain.NewUserID()}
	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestUpdateUserHandler_InvalidEmail(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id domain.UserID) (*domain.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	badEmail := "not-an-email"
	cmd := UpdateUserCommand{
		ID:    domain.UserID(user.ID()),
		Email: &badEmail,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}

func TestUpdateUserHandler_OnlyAttributes(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id domain.UserID) (*domain.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	cmd := UpdateUserCommand{
		ID:         domain.UserID(user.ID()),
		Attributes: map[string]string{"new_key": "new_val"},
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedUser.Attributes()["new_key"] != "new_val" {
		t.Error("expected attributes to be updated")
	}
}
