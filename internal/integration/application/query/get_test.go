package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/integration/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.IntegrationView
	views []*domain.IntegrationView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.IntegrationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrIntegrationNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	return nil, domain.ErrIntegrationNotFound
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.IntegrationView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	return nil, 0, m.err
}

func (m *errorReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	return nil, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetHandler_Handle(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.IntegrationView{
			ID:         id,
			Name:       "Slack",
			Type:       "messaging",
			APIKey:     "xoxb-key",
			WebhookURL: "https://hooks.slack.com",
			Enabled:    true,
			Config:     map[string]any{"channel": "#general"},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Name != "Slack" {
		t.Errorf("expected name Slack, got %s", result.Name)
	}
	if result.Type != "messaging" {
		t.Errorf("expected type messaging, got %s", result.Type)
	}
	if result.APIKey != "xoxb-key" {
		t.Errorf("expected apiKey xoxb-key, got %s", result.APIKey)
	}
	if !result.Enabled {
		t.Error("expected enabled true")
	}
	if result.Config["channel"] != "#general" {
		t.Errorf("expected config channel #general, got %v", result.Config["channel"])
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.IntegrationView{
			ID:         id,
			Name:       "SMTP",
			Type:       "email",
			APIKey:     "smtp-key",
			WebhookURL: "",
			Enabled:    false,
			Config:     map[string]any{"host": "smtp.example.com", "port": float64(587)},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.WebhookURL != "" {
		t.Errorf("expected empty webhookURL, got %s", result.WebhookURL)
	}
	if result.Enabled {
		t.Error("expected enabled false")
	}
	if result.Config["host"] != "smtp.example.com" {
		t.Errorf("expected config host smtp.example.com, got %v", result.Config["host"])
	}
}
