package query

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/iam/user/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFindSessionHandler_Success(t *testing.T) {
	t.Parallel()

	sessionID := uuid.New()
	userID := uuid.New()
	deviceID := uuid.New()
	now := time.Now()

	readRepo := &mockUserReadRepository{
		session: &shared.AuthSession{
			ID:               sessionID,
			UserID:           userID,
			DeviceID:         deviceID,
			RefreshTokenHash: "hashed_token",
			ExpiresAt:        now.Add(24 * time.Hour),
			Revoked:          false,
			LastActivity:     now,
		},
	}

	handler := NewFindSessionHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: domain.SessionID(sessionID)})
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected session, got nil")
	}

	if result.ID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, result.ID)
	}

	if result.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, result.UserID)
	}

	if result.DeviceID != deviceID {
		t.Errorf("expected device ID %s, got %s", deviceID, result.DeviceID)
	}

	if result.Revoked {
		t.Error("expected session to not be revoked")
	}
}

func TestFindSessionHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockUserReadRepository{}

	handler := NewFindSessionHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: domain.NewSessionID()})
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}

func TestFindSessionHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewFindSessionHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: domain.NewSessionID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
