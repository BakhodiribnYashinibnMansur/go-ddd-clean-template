package command

import (
	"context"
	"testing"

	"gct/internal/context/admin/integration/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockIntegrationRepo struct {
	saved   *domain.Integration
	updated *domain.Integration
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.Integration, error)
}

func (m *mockIntegrationRepo) Save(_ context.Context, e *domain.Integration) error {
	m.saved = e
	return nil
}

func (m *mockIntegrationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrIntegrationNotFound
}

func (m *mockIntegrationRepo) Update(_ context.Context, e *domain.Integration) error {
	m.updated = e
	return nil
}

func (m *mockIntegrationRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                                {}
func (m *mockLogger) Debugf(_ string, _ ...any)                     {}
func (m *mockLogger) Debugw(_ string, _ ...any)                     {}
func (m *mockLogger) Info(_ ...any)                                 {}
func (m *mockLogger) Infof(_ string, _ ...any)                      {}
func (m *mockLogger) Infow(_ string, _ ...any)                      {}
func (m *mockLogger) Warn(_ ...any)                                 {}
func (m *mockLogger) Warnf(_ string, _ ...any)                      {}
func (m *mockLogger) Warnw(_ string, _ ...any)                      {}
func (m *mockLogger) Error(_ ...any)                                {}
func (m *mockLogger) Errorf(_ string, _ ...any)                     {}
func (m *mockLogger) Errorw(_ string, _ ...any)                     {}
func (m *mockLogger) Fatal(_ ...any)                                {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                     {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)  {}

// --- Tests ---

func TestCreateHandler_Handle(t *testing.T) {
	repo := &mockIntegrationRepo{}
	eb := &mockEventBus{}
	handler := NewCreateHandler(repo, eb, &mockLogger{})

	cmd := CreateCommand{
		Name:       "Slack",
		Type:       "messaging",
		APIKey:     "xoxb-test-key",
		WebhookURL: "https://hooks.slack.com/test",
		Enabled:    true,
		Config:     map[string]string{"channel": "#general"},
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected integration to be saved")
	}
	if repo.saved.Name() != "Slack" {
		t.Errorf("expected name Slack, got %s", repo.saved.Name())
	}
	if repo.saved.Type() != "messaging" {
		t.Errorf("expected type messaging, got %s", repo.saved.Type())
	}
	if repo.saved.APIKey() != "xoxb-test-key" {
		t.Errorf("expected apiKey xoxb-test-key, got %s", repo.saved.APIKey())
	}
	if repo.saved.WebhookURL() != "https://hooks.slack.com/test" {
		t.Errorf("expected webhookURL, got %s", repo.saved.WebhookURL())
	}
	if !repo.saved.Enabled() {
		t.Error("expected enabled true")
	}
	if repo.saved.Config()["channel"] != "#general" {
		t.Errorf("expected config channel #general, got %v", repo.saved.Config()["channel"])
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "integration.connected" {
		t.Errorf("expected integration.connected, got %s", eb.published[0].EventName())
	}
}

func TestCreateHandler_NilConfig(t *testing.T) {
	repo := &mockIntegrationRepo{}
	handler := NewCreateHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateCommand{
		Name:       "SMTP",
		Type:       "email",
		APIKey:     "key",
		WebhookURL: "https://example.com",
		Enabled:    false,
		Config:     nil,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected integration to be saved")
	}
	if repo.saved.Config() == nil {
		t.Error("expected config to be initialized to empty map")
	}
}
