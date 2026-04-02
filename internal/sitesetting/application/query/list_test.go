package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

func TestListSiteSettingsHandler_Handle(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.SiteSettingView{
			{ID: uuid.New(), Key: "site_name", Value: "My Site", Type: "general", Description: "Name", CreatedAt: now, UpdatedAt: now},
			{ID: uuid.New(), Key: "maintenance", Value: "false", Type: "system", Description: "Maint", CreatedAt: now, UpdatedAt: now},
		},
		total: 2,
	}

	handler := NewListSiteSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: 10, Offset: 0},
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
	if result.Settings[0].Key != "site_name" {
		t.Errorf("expected site_name, got %s", result.Settings[0].Key)
	}
}

func TestListSiteSettingsHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.SiteSettingView{}, total: 0}

	handler := NewListSiteSettingsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{},
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

func TestListSiteSettingsHandler_WithFilters(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.SiteSettingView{
			{ID: uuid.New(), Key: "site_name", Value: "My Site", Type: "general", CreatedAt: now, UpdatedAt: now},
		},
		total: 1,
	}

	handler := NewListSiteSettingsHandler(readRepo, logger.Noop())
	settingType := "general"

	result, err := handler.Handle(context.Background(), ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{
			Type:  &settingType,
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListSiteSettingsHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListSiteSettingsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListSiteSettingsQuery{Filter: domain.SiteSettingFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
