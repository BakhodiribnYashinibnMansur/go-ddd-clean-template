package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/sitesetting/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.SiteSettingView
	views []*domain.SiteSettingView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.SiteSettingView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrSiteSettingNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.SiteSettingFilter) ([]*domain.SiteSettingView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.SiteSettingView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.SiteSettingFilter) ([]*domain.SiteSettingView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetSiteSetting ---

func TestGetSiteSettingHandler_Handle(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.SiteSettingView{
			ID:          id,
			Key:         "site_name",
			Value:       "My Site",
			Type:        "general",
			Description: "The name of the site",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := NewGetSiteSettingHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetSiteSettingQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Key != "site_name" {
		t.Errorf("expected key site_name, got %s", result.Key)
	}
	if result.Value != "My Site" {
		t.Errorf("expected value My Site, got %s", result.Value)
	}
	if result.Type != "general" {
		t.Errorf("expected type general, got %s", result.Type)
	}
	if result.Description != "The name of the site" {
		t.Errorf("expected description, got %s", result.Description)
	}
}

func TestGetSiteSettingHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetSiteSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetSiteSettingQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetSiteSettingHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetSiteSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetSiteSettingQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetSiteSettingHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	readRepo := &mockReadRepo{
		view: &domain.SiteSettingView{
			ID:          id,
			Key:         "maintenance_mode",
			Value:       "true",
			Type:        "system",
			Description: "Enable maintenance mode",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := NewGetSiteSettingHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetSiteSettingQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != id {
		t.Error("ID not mapped correctly")
	}
	if result.Key != "maintenance_mode" {
		t.Error("key not mapped")
	}
	if result.Value != "true" {
		t.Error("value not mapped")
	}
	if result.Type != "system" {
		t.Error("type not mapped")
	}
	if result.Description != "Enable maintenance mode" {
		t.Error("description not mapped")
	}
	if result.CreatedAt.IsZero() {
		t.Error("created_at not mapped")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("updated_at not mapped")
	}
}
