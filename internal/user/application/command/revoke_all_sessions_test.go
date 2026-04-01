package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// revokeTestRepo is a specialized mock for revoke_all_sessions tests that supports
// configurable Update and FindByID behavior.
type revokeTestRepo struct {
	mockUserRepository
	updateErr error
	updated   *domain.User
}

func (r *revokeTestRepo) Update(_ context.Context, entity *domain.User) error {
	r.updated = entity
	return r.updateErr
}

func TestRevokeAllSessionsHandler_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)
	user.Approve()

	// Add two sessions
	_, err := user.AddSession(domain.DeviceDesktop, "10.0.0.1", "Agent1")
	if err != nil {
		t.Fatalf("AddSession: %v", err)
	}
	_, err = user.AddSession(domain.DeviceMobile, "10.0.0.2", "Agent2")
	if err != nil {
		t.Fatalf("AddSession: %v", err)
	}

	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
				if id == user.ID() {
					return user, nil
				}
				return nil, domain.ErrUserNotFound
			},
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err = handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: user.ID(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	repo := &revokeTestRepo{} // default findByIDFn returns ErrUserNotFound
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRevokeAllSessionsHandler_UpdateError(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)

	updateErr := errors.New("db write failed")
	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
				if id == user.ID() {
					return user, nil
				}
				return nil, domain.ErrUserNotFound
			},
		},
		updateErr: updateErr,
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: user.ID(),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestRevokeAllSessionsHandler_NoSessions(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)

	repo := &revokeTestRepo{
		mockUserRepository: mockUserRepository{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
				if id == user.ID() {
					return user, nil
				}
				return nil, domain.ErrUserNotFound
			},
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}

	handler := NewRevokeAllSessionsHandler(repo, eb, l)

	err := handler.Handle(context.Background(), RevokeAllSessionsCommand{
		UserID: user.ID(),
	})
	if err != nil {
		t.Fatalf("expected no error for user with no sessions, got %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected user to be updated")
	}
	if len(repo.updated.Sessions()) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(repo.updated.Sessions()))
	}
}
