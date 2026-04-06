package command

import (
	"context"
	"errors"
	"testing"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/stretchr/testify/require"
)

// revokeTestRepo is a specialized mock for revoke_all_sessions tests that supports
// configurable Update and FindByID behavior.
type revokeTestRepo struct {
	mockUserRepository
	updateErr error
	updated   *userentity.User
}

func (r *revokeTestRepo) Update(_ context.Context, entity *userentity.User) error {
	r.updated = entity
	return r.updateErr
}

func TestRevokeAllSessionsHandler_Success(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, pw)
	user.Approve()

	// Add two sessions
	_, err := user.AddSession(userentity.DeviceDesktop, "10.0.0.1", "Agent1", "gct-client")
	require.NoError(t, err)
	_, err = user.AddSession(userentity.DeviceMobile, "10.0.0.2", "Agent2", "gct-client")
	require.NoError(t, err)

	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
				if id == user.TypedID() {
					return user, nil
				}
				return nil, userentity.ErrUserNotFound
			},
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err = handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: userentity.UserID(user.ID()),
	})
	require.NoError(t, err)

	if repo.updated == nil {
		t.Fatal("expected user to be updated")
	}

	for i, s := range repo.updated.Sessions() {
		if !s.IsRevoked() {
			t.Errorf("session %d should be revoked", i)
		}
	}
}

func TestRevokeAllSessionsHandler_UserNotFound(t *testing.T) {
	t.Parallel()

	repo := &revokeTestRepo{} // default findByIDFn returns ErrUserNotFound
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: userentity.NewUserID(),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, userentity.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRevokeAllSessionsHandler_UpdateError(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, pw)

	updateErr := errors.New("db write failed")
	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
				if id == user.TypedID() {
					return user, nil
				}
				return nil, userentity.ErrUserNotFound
			},
		},
		updateErr: updateErr,
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: userentity.UserID(user.ID()),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestRevokeAllSessionsHandler_NoSessions(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	user, _ := userentity.NewUser(phone, pw)

	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
				if id == user.TypedID() {
					return user, nil
				}
				return nil, userentity.ErrUserNotFound
			},
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: userentity.UserID(user.ID()),
	})
	require.NoError(t, err)
	if repo.updated == nil {
		t.Fatal("expected user to be updated")
	}
	if len(repo.updated.Sessions()) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(repo.updated.Sessions()))
	}
}
