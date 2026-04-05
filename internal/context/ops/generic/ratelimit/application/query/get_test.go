package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/ops/generic/ratelimit/domain"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.RateLimitView
	views []*domain.RateLimitView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id domain.RateLimitID) (*domain.RateLimitView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrRateLimitNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.RateLimitFilter) ([]*domain.RateLimitView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.RateLimitID) (*domain.RateLimitView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.RateLimitFilter) ([]*domain.RateLimitView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetRateLimitHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewRateLimitID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.RateLimitView{
			ID:                id,
			Name:              "api-global",
			Rule:              "/api/*",
			RequestsPerWindow: 100,
			WindowDuration:    60,
			Enabled:           true,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	handler := NewGetRateLimitHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetRateLimitQuery{ID: id})
	require.NoError(t, err)
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Name != "api-global" {
		t.Errorf("expected name api-global, got %s", result.Name)
	}
	if result.RequestsPerWindow != 100 {
		t.Errorf("expected requestsPerWindow 100, got %d", result.RequestsPerWindow)
	}
	if result.WindowDuration != 60 {
		t.Errorf("expected windowDuration 60, got %d", result.WindowDuration)
	}
	if !result.Enabled {
		t.Error("expected enabled true")
	}
}

func TestGetRateLimitHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetRateLimitHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRateLimitQuery{ID: domain.NewRateLimitID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetRateLimitHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetRateLimitHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRateLimitQuery{ID: domain.NewRateLimitID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetRateLimitHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := domain.NewRateLimitID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.RateLimitView{
			ID:                id,
			Name:              "strict",
			Rule:              "/auth/*",
			RequestsPerWindow: 5,
			WindowDuration:    10,
			Enabled:           false,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	handler := NewGetRateLimitHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetRateLimitQuery{ID: id})
	require.NoError(t, err)
	if result.Rule != "/auth/*" {
		t.Errorf("expected rule /auth/*, got %s", result.Rule)
	}
	if result.Enabled {
		t.Error("expected enabled false")
	}
}
