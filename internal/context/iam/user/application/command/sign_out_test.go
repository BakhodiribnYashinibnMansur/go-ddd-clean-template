package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
)

func TestSignOutHandler_Handle(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)
	user.Approve()

	session, err := user.AddSession(domain.DeviceDesktop, "10.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("AddSession: %v", err)
	}

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

	handler := NewSignOutHandler(repo, eventBus, log)

	err = handler.Handle(context.Background(), SignOutCommand{
		UserID:    domain.UserID(user.ID()),
		SessionID: domain.SessionID(session.ID()),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	user := domain.NewUser(phone, pw)

	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignOutHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), SignOutCommand{
		UserID:    domain.UserID(user.ID()),
		SessionID: domain.NewSessionID(), // non-existent session
	})
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSignOutHandler_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignOutHandler(repo, eventBus, log)

	err := handler.Handle(context.Background(), SignOutCommand{
		UserID:    domain.NewUserID(),
		SessionID: domain.NewSessionID(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
