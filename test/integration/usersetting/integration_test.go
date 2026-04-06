package usersetting

import (
	"context"
	"testing"

	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/iam/generic/usersetting"
	"gct/internal/context/iam/generic/usersetting/application/command"
	"gct/internal/context/iam/generic/usersetting/application/query"
	settingentity "gct/internal/context/iam/generic/usersetting/domain/entity"
	settingrepo "gct/internal/context/iam/generic/usersetting/domain/repository"
	"gct/test/integration/common/setup"

	"github.com/google/uuid"
)

func newTestBC(t *testing.T) *usersetting.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return usersetting.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetUserSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.UpsertUserSetting.Handle(ctx, command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "theme",
		Value:  "dark",
	})
	if err != nil {
		t.Fatalf("UpsertUserSetting: %v", err)
	}

	result, err := bc.ListUserSettings.Handle(ctx, query.ListUserSettingsQuery{
		Filter: settingrepo.UserSettingFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListUserSettings: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 user setting, got %d", result.Total)
	}

	s := result.Settings[0]
	if s.Key != "theme" {
		t.Errorf("expected key theme, got %s", s.Key)
	}
	if s.Value != "dark" {
		t.Errorf("expected value dark, got %s", s.Value)
	}

	view, err := bc.GetUserSetting.Handle(ctx, query.GetUserSettingQuery{ID: settingentity.UserSettingID(s.ID)})
	if err != nil {
		t.Fatalf("GetUserSetting: %v", err)
	}
	if view.ID != s.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, s.ID)
	}
}

func TestIntegration_UpsertUserSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.UpsertUserSetting.Handle(ctx, command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "language",
		Value:  "en",
	})
	if err != nil {
		t.Fatalf("UpsertUserSetting (create): %v", err)
	}

	// Upsert again with a new value for the same key
	err = bc.UpsertUserSetting.Handle(ctx, command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "language",
		Value:  "uz",
	})
	if err != nil {
		t.Fatalf("UpsertUserSetting (update): %v", err)
	}

	result, err := bc.ListUserSettings.Handle(ctx, query.ListUserSettingsQuery{
		Filter: settingrepo.UserSettingFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListUserSettings: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 user setting after upsert, got %d", result.Total)
	}

	if result.Settings[0].Value != "uz" {
		t.Errorf("expected value uz after upsert, got %s", result.Settings[0].Value)
	}
}

func TestIntegration_DeleteUserSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.UpsertUserSetting.Handle(ctx, command.UpsertUserSettingCommand{
		UserID: userID,
		Key:    "timezone",
		Value:  "UTC+5",
	})
	if err != nil {
		t.Fatalf("UpsertUserSetting: %v", err)
	}

	list, _ := bc.ListUserSettings.Handle(ctx, query.ListUserSettingsQuery{
		Filter: settingrepo.UserSettingFilter{Limit: 10},
	})
	sID := settingentity.UserSettingID(list.Settings[0].ID)

	err = bc.DeleteUserSetting.Handle(ctx, command.DeleteUserSettingCommand{ID: sID})
	if err != nil {
		t.Fatalf("DeleteUserSetting: %v", err)
	}

	list2, _ := bc.ListUserSettings.Handle(ctx, query.ListUserSettingsQuery{
		Filter: settingrepo.UserSettingFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 user settings after delete, got %d", list2.Total)
	}
}
