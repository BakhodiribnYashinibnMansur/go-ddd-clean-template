package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"testing"
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

func TestFindSessionHandler_Success(t *testing.T) {
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

	result, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: sessionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	readRepo := &mockUserReadRepository{}

	handler := NewFindSessionHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}

func TestFindSessionHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewFindSessionHandler(readRepo, logger.Noop())

	_, err := handler.Handle(context.Background(), FindSessionQuery{SessionID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
