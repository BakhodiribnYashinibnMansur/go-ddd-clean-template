package command

import (
	"context"
	"errors"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestSignOutHandler_Handle(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, pw)
	user.Approve()

	session, err := user.AddSession(userentity.DeviceDesktop, "10.0.0.1", "TestAgent", "gct-client")
	require.NoError(t, err)

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

	handler := NewSignOutHandler(repo, fakeDB{}, eventBus, log)

	err = handler.Handle(context.Background(), SignOutCommand{
		UserID:    userentity.UserID(user.ID()),
		SessionID: userentity.SessionID(session.ID()),
	})
	require.NoError(t, err)

	if repo.updatedUser == nil {
		t.Fatal("expected user to be updated")
	}

	if len(repo.updatedUser.Sessions()) != 1 {
		t.Errorf("expected 1 session after sign-out (revoked), got %d", len(repo.updatedUser.Sessions()))
	}
	if !repo.updatedUser.Sessions()[0].IsRevoked() {
		t.Error("session should be revoked after sign-out")
	}
}

func TestSignOutHandler_SessionNotFound(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignOutHandler(repo, fakeDB{}, eventBus, log)

	err := handler.Handle(context.Background(), SignOutCommand{
		UserID:    userentity.UserID(user.ID()),
		SessionID: userentity.NewSessionID(), // non-existent session
	})
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
	if !errors.Is(err, userentity.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSignOutHandler_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignOutHandler(repo, fakeDB{}, eventBus, log)

	err := handler.Handle(context.Background(), SignOutCommand{
		UserID:    userentity.NewUserID(),
		SessionID: userentity.NewSessionID(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
