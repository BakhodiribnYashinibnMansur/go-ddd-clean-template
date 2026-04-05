package query

import (
	"context"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/iam/generic/usersetting/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListUserSettingsHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.UserSettingView{
			{ID: domain.NewUserSettingID(), UserID: uuid.New(), Key: "theme", Value: "dark", CreatedAt: now, UpdatedAt: now},
			{ID: domain.NewUserSettingID(), UserID: uuid.New(), Key: "locale", Value: "en", CreatedAt: now, UpdatedAt: now},
		},
		total: 2,
	}

	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListUserSettingsQuery{
		Filter: domain.UserSettingFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{views: []*domain.UserSettingView{}, total: 0}

	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListUserSettingsQuery{
		Filter: domain.UserSettingFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Settings) != 0 {
		t.Errorf("expected 0 settings, got %d", len(result.Settings))
	}
}

func TestListUserSettingsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListUserSettingsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListUserSettingsQuery{Filter: domain.UserSettingFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
