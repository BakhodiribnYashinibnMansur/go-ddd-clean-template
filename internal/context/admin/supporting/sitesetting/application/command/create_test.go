package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/sitesetting/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.SiteSetting
	updated *domain.SiteSetting
	findFn  func(ctx context.Context, id domain.SiteSettingID) (*domain.SiteSetting, error)
}

func (m *mockRepo) Save(_ context.Context, e *domain.SiteSetting) error {
	m.saved = e
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id domain.SiteSettingID) (*domain.SiteSetting, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrSiteSettingNotFound
}

func (m *mockRepo) Update(_ context.Context, e *domain.SiteSetting) error {
	m.updated = e
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _ domain.SiteSettingID) error {
	return nil
}

func (m *mockRepo) List(_ context.Context, _ domain.SiteSettingFilter) ([]*domain.SiteSetting, int64, error) {
	return nil, 0, nil
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

func (m *mockLogger) Debug(args ...any)                            {}
func (m *mockLogger) Debugf(template string, args ...any)          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Info(args ...any)                             {}
func (m *mockLogger) Infof(template string, args ...any)           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Warn(args ...any)                             {}
func (m *mockLogger) Warnf(template string, args ...any)           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Error(args ...any)                            {}
func (m *mockLogger) Errorf(template string, args ...any)          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Fatal(args ...any)                            {}
func (m *mockLogger) Fatalf(template string, args ...any)          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// --- Tests ---

func TestCreateSiteSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSiteSettingHandler(repo, eb, log)

	cmd := CreateSiteSettingCommand{
		Key:         "site_name",
		Value:       "My Site",
		Type:        "general",
		Description: "The name of the site",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected site setting to be saved")
	}
	if repo.saved.Key() != "site_name" {
		t.Errorf("expected key site_name, got %s", repo.saved.Key())
	}
	if repo.saved.Value() != "My Site" {
		t.Errorf("expected value My Site, got %s", repo.saved.Value())
	}
	if repo.saved.Type() != "general" {
		t.Errorf("expected type general, got %s", repo.saved.Type())
	}
	if repo.saved.Description() != "The name of the site" {
		t.Errorf("expected description, got %s", repo.saved.Description())
	}
}

func TestCreateSiteSettingHandler_MinimalFields(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSiteSettingHandler(repo, eb, log)

	err := handler.Handle(context.Background(), CreateSiteSettingCommand{
		Key:   "maintenance_mode",
		Value: "false",
		Type:  "system",
	})
	require.NoError(t, err)
	if repo.saved == nil {
		t.Fatal("expected site setting to be saved")
	}
	if repo.saved.Key() != "maintenance_mode" {
		t.Errorf("expected key maintenance_mode, got %s", repo.saved.Key())
	}
}

func TestCreateSiteSettingHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo save failed")
	errR := &errorRepo{saveErr: repoErr}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSiteSettingHandler(errR, eb, log)
	err := handler.Handle(context.Background(), CreateSiteSettingCommand{
		Key: "k", Value: "v", Type: "t",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo save error, got: %v", err)
	}
}
