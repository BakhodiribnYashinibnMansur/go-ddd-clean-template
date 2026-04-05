package application

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/integration/domain"
	"gct/internal/kernel/consts"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type cacheTestReadRepo struct {
	views []*domain.IntegrationView
	total int64
	err   error
}

func (m *cacheTestReadRepo) FindByID(_ context.Context, _ domain.IntegrationID) (*domain.IntegrationView, error) {
	return nil, domain.ErrIntegrationNotFound
}

func (m *cacheTestReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.views, m.total, nil
}

func (m *cacheTestReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	return nil, domain.ErrIntegrationNotFound
}

type cacheTestLogger struct{}

func (m *cacheTestLogger) Debug(args ...any)                            {}
func (m *cacheTestLogger) Debugf(template string, args ...any)          {}
func (m *cacheTestLogger) Debugw(msg string, keysAndValues ...any)      {}
func (m *cacheTestLogger) Info(args ...any)                             {}
func (m *cacheTestLogger) Infof(template string, args ...any)           {}
func (m *cacheTestLogger) Infow(msg string, keysAndValues ...any)       {}
func (m *cacheTestLogger) Warn(args ...any)                             {}
func (m *cacheTestLogger) Warnf(template string, args ...any)           {}
func (m *cacheTestLogger) Warnw(msg string, keysAndValues ...any)       {}
func (m *cacheTestLogger) Error(args ...any)                            {}
func (m *cacheTestLogger) Errorf(template string, args ...any)          {}
func (m *cacheTestLogger) Errorw(msg string, keysAndValues ...any)      {}
func (m *cacheTestLogger) Fatal(args ...any)                            {}
func (m *cacheTestLogger) Fatalf(template string, args ...any)          {}
func (m *cacheTestLogger) Fatalw(msg string, keysAndValues ...any)      {}
func (m *cacheTestLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *cacheTestLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *cacheTestLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *cacheTestLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *cacheTestLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// ---------------------------------------------------------------------------
// Tests: InitCache
// ---------------------------------------------------------------------------

func TestCacheService_InitCache_Success(t *testing.T) {
	t.Parallel()

	id1 := domain.NewIntegrationID()
	id2 := domain.NewIntegrationID()
	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{
			{
				ID:      id1,
				Name:    "Slack",
				Type:    "messaging",
				APIKey:  "xoxb-key-1",
				Enabled: true,
				Config:  map[string]string{"channel": "#general"},
			},
			{
				ID:      id2,
				Name:    "SMTP",
				Type:    "email",
				APIKey:  "smtp-key-2",
				Enabled: true,
				Config:  map[string]string{"host": "smtp.example.com"},
			},
		},
		total: 2,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InitCache(context.Background())
	require.NoError(t, err)

	// Verify FindByID
	ci, ok := svc.FindByID(id1)
	if !ok {
		t.Fatal("expected to find integration by ID")
	}
	if ci.Name != "Slack" {
		t.Errorf("expected name Slack, got %s", ci.Name)
	}
	if ci.APIKey != "xoxb-key-1" {
		t.Errorf("expected API key xoxb-key-1, got %s", ci.APIKey)
	}

	ci2, ok := svc.FindByID(id2)
	if !ok {
		t.Fatal("expected to find second integration by ID")
	}
	if ci2.Name != "SMTP" {
		t.Errorf("expected name SMTP, got %s", ci2.Name)
	}

	// Verify FindByAPIKey
	ci, ok = svc.FindByAPIKey("xoxb-key-1")
	if !ok {
		t.Fatal("expected to find integration by API key")
	}
	if ci.ID != id1 {
		t.Errorf("expected ID %s, got %s", id1, ci.ID)
	}

	ci2, ok = svc.FindByAPIKey("smtp-key-2")
	if !ok {
		t.Fatal("expected to find second integration by API key")
	}
	if ci2.ID != id2 {
		t.Errorf("expected ID %s, got %s", id2, ci2.ID)
	}
}

func TestCacheService_InitCache_Empty(t *testing.T) {
	t.Parallel()

	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{},
		total: 0,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InitCache(context.Background())
	require.NoError(t, err)

	_, ok := svc.FindByID(domain.NewIntegrationID())
	if ok {
		t.Error("expected not found for random ID on empty cache")
	}

	_, ok = svc.FindByAPIKey("nonexistent")
	if ok {
		t.Error("expected not found for random key on empty cache")
	}
}

func TestCacheService_InitCache_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("database unavailable")
	repo := &cacheTestReadRepo{err: repoErr}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InitCache(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestCacheService_InitCache_EmptyAPIKeySkipped(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{
			{
				ID:      id,
				Name:    "WebhookOnly",
				Type:    "webhook",
				APIKey:  "", // no API key
				Enabled: true,
			},
		},
		total: 1,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InitCache(context.Background())
	require.NoError(t, err)

	// Should be findable by ID
	ci, ok := svc.FindByID(id)
	if !ok {
		t.Fatal("expected to find integration by ID")
	}
	if ci.Name != "WebhookOnly" {
		t.Errorf("expected name WebhookOnly, got %s", ci.Name)
	}

	// Should NOT be in apiKeys map (empty key)
	_, ok = svc.FindByAPIKey("")
	if ok {
		t.Error("expected empty API key to not be cached")
	}
}

// ---------------------------------------------------------------------------
// Tests: FindByAPIKey / FindByID
// ---------------------------------------------------------------------------

func TestCacheService_FindByAPIKey_NotFound(t *testing.T) {
	t.Parallel()

	repo := &cacheTestReadRepo{views: []*domain.IntegrationView{}}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)
	_ = svc.InitCache(context.Background())

	_, ok := svc.FindByAPIKey("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestCacheService_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := &cacheTestReadRepo{views: []*domain.IntegrationView{}}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)
	_ = svc.InitCache(context.Background())

	_, ok := svc.FindByID(domain.NewIntegrationID())
	if ok {
		t.Error("expected not found")
	}
}

// ---------------------------------------------------------------------------
// Tests: InvalidateCache
// ---------------------------------------------------------------------------

func TestCacheService_InvalidateCache_IntegrationsTable(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{
			{ID: id, Name: "Slack", APIKey: "key-1", Enabled: true},
		},
		total: 1,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InvalidateCache(context.Background(), consts.TableIntegrations)
	require.NoError(t, err)

	// After invalidation, cache should have data from the repo
	ci, ok := svc.FindByID(id)
	if !ok {
		t.Fatal("expected integration to be cached after invalidation")
	}
	if ci.Name != "Slack" {
		t.Errorf("expected name Slack, got %s", ci.Name)
	}
}

func TestCacheService_InvalidateCache_APIKeysTable(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{
			{ID: id, Name: "SMTP", APIKey: "smtp-key", Enabled: true},
		},
		total: 1,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InvalidateCache(context.Background(), consts.TableAPIKeys)
	require.NoError(t, err)

	ci, ok := svc.FindByAPIKey("smtp-key")
	if !ok {
		t.Fatal("expected integration to be cached after api_keys invalidation")
	}
	if ci.Name != "SMTP" {
		t.Errorf("expected name SMTP, got %s", ci.Name)
	}
}

func TestCacheService_InvalidateCache_UnrelatedTable(t *testing.T) {
	t.Parallel()

	repo := &cacheTestReadRepo{
		views: []*domain.IntegrationView{},
		total: 0,
	}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	// Unrelated table should not trigger re-init
	err := svc.InvalidateCache(context.Background(), "some_other_table")
	require.NoError(t, err)
}

func TestCacheService_InvalidateCache_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db connection lost")
	repo := &cacheTestReadRepo{err: repoErr}
	l := &cacheTestLogger{}
	svc := NewCacheService(repo, l)

	err := svc.InvalidateCache(context.Background(), consts.TableIntegrations)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
