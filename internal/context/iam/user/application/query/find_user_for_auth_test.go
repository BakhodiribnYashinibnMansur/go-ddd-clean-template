package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"testing"

	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

func TestFindUserForAuthHandler_Success(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	readRepo := &mockUserReadRepository{
		authUser: &shared.AuthUser{
			ID:         userID,
			RoleID:     &roleID,
			Active:     true,
			IsApproved: true,
			Attributes: map[string]string{"level": "10"},
		},
	}

	handler := NewFindUserForAuthHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: userID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected auth user, got nil")
	}

	if result.ID != userID {
		t.Errorf("expected user ID %s, got %s", userID, result.ID)
	}

	if result.RoleID == nil || *result.RoleID != roleID {
		t.Error("expected roleID to be set")
	}

	if !result.Active {
		t.Error("expected user to be active")
	}

	if !result.IsApproved {
		t.Error("expected user to be approved")
	}

	if result.Attributes["level"] != "10" {
		t.Error("expected attributes to be mapped")
	}
}

func TestFindUserForAuthHandler_NotFound(t *testing.T) {
	readRepo := &mockUserReadRepository{}

	handler := NewFindUserForAuthHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestFindUserForAuthHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewFindUserForAuthHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
