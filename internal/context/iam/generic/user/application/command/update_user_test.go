package command

import (
	"context"
	"testing"
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func makeTestUser(t *testing.T) *userentity.User {
	t.Helper()
	phone, err := userentity.NewPhone("+998901234567")
	require.NoError(t, err)
	pw, err := userentity.NewPasswordFromRaw("StrongP@ss123")
	require.NoError(t, err)
	email, _ := userentity.NewEmail("old@example.com")
	username := "olduser"
	return userentity.ReconstructUser(
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
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == user.TypedID() {
				return user, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	newEmail := "new@example.com"
	newUsername := "newuser"
	cmd := UpdateUserCommand{
		ID:       userentity.UserID(user.ID()),
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

	cmd := UpdateUserCommand{ID: userentity.NewUserID()}
	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestUpdateUserHandler_InvalidEmail(t *testing.T) {
	t.Parallel()

	user := makeTestUser(t)
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	badEmail := "not-an-email"
	cmd := UpdateUserCommand{
		ID:    userentity.UserID(user.ID()),
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
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateUserHandler(repo, eventBus, log)

	cmd := UpdateUserCommand{
		ID:         userentity.UserID(user.ID()),
		Attributes: map[string]string{"new_key": "new_val"},
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedUser.Attributes()["new_key"] != "new_val" {
		t.Error("expected attributes to be updated")
	}
}
