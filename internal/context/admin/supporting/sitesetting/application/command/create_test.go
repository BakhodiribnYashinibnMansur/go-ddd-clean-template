package command

import (
	"context"
	"errors"
	"testing"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockRepo struct {
	saved   *siteentity.SiteSetting
	updated *siteentity.SiteSetting
	findFn  func(ctx context.Context, id siteentity.SiteSettingID) (*siteentity.SiteSetting, error)
}

func (m *mockRepo) Save(_ context.Context, _ shareddomain.Querier, e *siteentity.SiteSetting) error {
	m.saved = e
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id siteentity.SiteSettingID) (*siteentity.SiteSetting, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, siteentity.ErrSiteSettingNotFound
}

func (m *mockRepo) Update(_ context.Context, _ shareddomain.Querier, e *siteentity.SiteSetting) error {
	m.updated = e
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _ shareddomain.Querier, _ siteentity.SiteSettingID) error {
	return nil
}

func (m *mockRepo) List(_ context.Context, _ siterepo.SiteSettingFilter) ([]*siteentity.SiteSetting, int64, error) {
	return nil, 0, nil
}

type mockEventBus struct {
	published []shareddomain.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shareddomain.DomainEvent) error {
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

	handler := NewCreateSiteSettingHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)

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

	handler := NewCreateSiteSettingHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)

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

	handler := NewCreateSiteSettingHandler(errR, outbox.NewEventCommitter(nil, nil, eb, log), log)
	err := handler.Handle(context.Background(), CreateSiteSettingCommand{
		Key: "k", Value: "v", Type: "t",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo save error, got: %v", err)
	}
}
