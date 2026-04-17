package command

import (
	"context"
	"testing"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockRateLimitRepo struct {
	saved   *ratelimitentity.RateLimit
	updated *ratelimitentity.RateLimit
	deleted ratelimitentity.RateLimitID
	findFn  func(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error)
}

func (m *mockRateLimitRepo) Save(_ context.Context, _ shared.Querier, e *ratelimitentity.RateLimit) error {
	m.saved = e
	return nil
}

func (m *mockRateLimitRepo) FindByID(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, ratelimitentity.ErrRateLimitNotFound
}

func (m *mockRateLimitRepo) Update(_ context.Context, _ shared.Querier, e *ratelimitentity.RateLimit) error {
	m.updated = e
	return nil
}

func (m *mockRateLimitRepo) Delete(_ context.Context, _ shared.Querier, id ratelimitentity.RateLimitID) error {
	m.deleted = id
	return nil
}

func (m *mockRateLimitRepo) List(_ context.Context, _ ratelimitrepo.RateLimitFilter) ([]*ratelimitentity.RateLimit, int64, error) {
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

func (m *mockLogger) Debug(_ ...any)                               {}
func (m *mockLogger) Debugf(_ string, _ ...any)                    {}
func (m *mockLogger) Debugw(_ string, _ ...any)                    {}
func (m *mockLogger) Info(_ ...any)                                {}
func (m *mockLogger) Infof(_ string, _ ...any)                     {}
func (m *mockLogger) Infow(_ string, _ ...any)                     {}
func (m *mockLogger) Warn(_ ...any)                                {}
func (m *mockLogger) Warnf(_ string, _ ...any)                     {}
func (m *mockLogger) Warnw(_ string, _ ...any)                     {}
func (m *mockLogger) Error(_ ...any)                               {}
func (m *mockLogger) Errorf(_ string, _ ...any)                    {}
func (m *mockLogger) Errorw(_ string, _ ...any)                    {}
func (m *mockLogger) Fatal(_ ...any)                               {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                    {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// --- Tests ---

func TestCreateRateLimitHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRateLimitRepo{}
	eb := &mockEventBus{}
	handler := NewCreateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	cmd := CreateRateLimitCommand{
		Name:              "api-global",
		Rule:              "/api/*",
		RequestsPerWindow: 100,
		WindowDuration:    60,
		Enabled:           true,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected rate limit to be saved")
	}
	if repo.saved.Name() != "api-global" {
		t.Errorf("expected name api-global, got %s", repo.saved.Name())
	}
	if repo.saved.Rule() != "/api/*" {
		t.Errorf("expected rule /api/*, got %s", repo.saved.Rule())
	}
	if repo.saved.RequestsPerWindow() != 100 {
		t.Errorf("expected requestsPerWindow 100, got %d", repo.saved.RequestsPerWindow())
	}
	if repo.saved.WindowDuration() != 60 {
		t.Errorf("expected windowDuration 60, got %d", repo.saved.WindowDuration())
	}
	if !repo.saved.Enabled() {
		t.Error("expected enabled true")
	}
}

func TestCreateRateLimitHandler_Disabled(t *testing.T) {
	t.Parallel()

	repo := &mockRateLimitRepo{}
	handler := NewCreateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), CreateRateLimitCommand{
		Name:              "disabled-rule",
		Rule:              "/test",
		RequestsPerWindow: 10,
		WindowDuration:    30,
		Enabled:           false,
	})
	require.NoError(t, err)
	if repo.saved == nil {
		t.Fatal("expected rate limit to be saved")
	}
	if repo.saved.Enabled() {
		t.Error("expected enabled false")
	}
}
