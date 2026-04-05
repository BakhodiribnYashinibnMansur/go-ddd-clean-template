package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/integration/domain"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view       *domain.IntegrationView
	views      []*domain.IntegrationView
	total      int64
	apiKeyView *domain.IntegrationAPIKeyView
}

func (m *mockReadRepo) FindByID(_ context.Context, id domain.IntegrationID) (*domain.IntegrationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrIntegrationNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	if m.apiKeyView != nil {
		return m.apiKeyView, nil
	}
	return nil, domain.ErrIntegrationNotFound
}

func (m *mockReadRepo) ListActiveJWT(_ context.Context) ([]domain.JWTIntegrationView, error) {
	return nil, nil
}

func (m *mockReadRepo) FindJWTByHash(_ context.Context, _ []byte) (*domain.JWTIntegrationView, error) {
	return nil, domain.ErrIntegrationNotFound
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.IntegrationID) (*domain.IntegrationView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	return nil, 0, m.err
}

func (m *errorReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	return nil, m.err
}

func (m *errorReadRepo) ListActiveJWT(_ context.Context) ([]domain.JWTIntegrationView, error) {
	return nil, m.err
}

func (m *errorReadRepo) FindJWTByHash(_ context.Context, _ []byte) (*domain.JWTIntegrationView, error) {
	return nil, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.IntegrationView{
			ID:         id,
			Name:       "Slack",
			Type:       "messaging",
			APIKey:     "xoxb-key",
			WebhookURL: "https://hooks.slack.com",
			Enabled:    true,
			Config:     map[string]string{"channel": "#general"},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: domain.IntegrationID(id)})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetQuery{ID: domain.NewIntegrationID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetQuery{ID: domain.NewIntegrationID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.IntegrationView{
			ID:         id,
			Name:       "SMTP",
			Type:       "email",
			APIKey:     "smtp-key",
			WebhookURL: "",
			Enabled:    false,
			Config:     map[string]string{"host": "smtp.example.com", "port": "587"},
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: domain.IntegrationID(id)})
	require.NoError(t, err)
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
