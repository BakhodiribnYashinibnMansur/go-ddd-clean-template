package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/context/iam/usersetting/domain"

	"github.com/google/uuid"
)

func TestListUserSettingsHandler_Handle(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.UserSettingView{
			{ID: uuid.New(), UserID: uuid.New(), Key: "theme", Value: "dark", CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), UserID: uuid.New(), Key: "locale", Value: "en", CreatedAt: now, UpdatedAt: now},
		},
		total: 2,
	}

	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListUserSettingsQuery{
		Filter: domain.UserSettingFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Settings) != 2 {
		t.Fatalf("expected 2 settings, got %d", len(result.Settings))
	}
	if result.Settings[0].Key != "theme" {
		t.Errorf("expected 'theme', got %s", result.Settings[0].Key)
	}
}

func TestListUserSettingsHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.UserSettingView{}, total: 0}

	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListUserSettingsQuery{
		Filter: domain.UserSettingFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Settings) != 0 {
		t.Errorf("expected 0 settings, got %d", len(result.Settings))
	}
}

func TestListUserSettingsHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListUserSettingsQuery{Filter: domain.UserSettingFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
