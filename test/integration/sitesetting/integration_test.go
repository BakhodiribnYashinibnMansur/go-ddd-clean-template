package sitesetting

import (
	"context"
	"testing"

	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/sitesetting"
	"gct/internal/sitesetting/application/command"
	"gct/internal/sitesetting/application/query"
	"gct/internal/sitesetting/domain"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *sitesetting.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return sitesetting.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetSiteSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateSiteSetting.Handle(ctx, command.CreateSiteSettingCommand{
		Key:         "site_name",
		Value:       "My Application",
		Type:        "string",
		Description: "The name of the site",
	})
	if err != nil {
		t.Fatalf("CreateSiteSetting: %v", err)
	}

	result, err := bc.ListSiteSettings.Handle(ctx, query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSiteSettings: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 site setting, got %d", result.Total)
	}

	s := result.Settings[0]
	if s.Key != "site_name" {
		t.Errorf("expected key site_name, got %s", s.Key)
	}
	if s.Value != "My Application" {
		t.Errorf("expected value 'My Application', got %s", s.Value)
	}

	view, err := bc.GetSiteSetting.Handle(ctx, query.GetSiteSettingQuery{ID: s.ID})
	if err != nil {
		t.Fatalf("GetSiteSetting: %v", err)
	}
	if view.ID != s.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, s.ID)
	}
}

func TestIntegration_UpdateSiteSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateSiteSetting.Handle(ctx, command.CreateSiteSettingCommand{
		Key:         "maintenance_mode",
		Value:       "false",
		Type:        "boolean",
		Description: "Toggle maintenance mode",
	})
	if err != nil {
		t.Fatalf("CreateSiteSetting: %v", err)
	}

	list, _ := bc.ListSiteSettings.Handle(ctx, query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: 10},
	})
	sID := list.Settings[0].ID

	newValue := "true"
	err = bc.UpdateSiteSetting.Handle(ctx, command.UpdateSiteSettingCommand{
		ID:    sID,
		Value: &newValue,
	})
	if err != nil {
		t.Fatalf("UpdateSiteSetting: %v", err)
	}

	view, _ := bc.GetSiteSetting.Handle(ctx, query.GetSiteSettingQuery{ID: sID})
	if view.Value != "true" {
		t.Errorf("value not updated, got %s", view.Value)
	}
}

func TestIntegration_DeleteSiteSetting(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateSiteSetting.Handle(ctx, command.CreateSiteSettingCommand{
		Key:         "delete_me",
		Value:       "temporary",
		Type:        "string",
		Description: "Will be deleted",
	})
	if err != nil {
		t.Fatalf("CreateSiteSetting: %v", err)
	}

	list, _ := bc.ListSiteSettings.Handle(ctx, query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: 10},
	})
	sID := list.Settings[0].ID

	err = bc.DeleteSiteSetting.Handle(ctx, command.DeleteSiteSettingCommand{ID: sID})
	if err != nil {
		t.Fatalf("DeleteSiteSetting: %v", err)
	}

	list2, _ := bc.ListSiteSettings.Handle(ctx, query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 site settings after delete, got %d", list2.Total)
	}
}
