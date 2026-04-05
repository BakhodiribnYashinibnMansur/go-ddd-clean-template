package query

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/user/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFindUserForAuthHandler_Success(t *testing.T) {
	t.Parallel()

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

	result, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: domain.UserID(userID)})
	require.NoError(t, err)

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
	t.Parallel()

	readRepo := &mockUserReadRepository{}

	handler := NewFindUserForAuthHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: domain.NewUserID()})
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestFindUserForAuthHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewFindUserForAuthHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindUserForAuthQuery{UserID: domain.NewUserID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
