package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/context/ops/ratelimit/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListRateLimitsHandler_Handle(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*domain.RateLimitView{
			{ID: uuid.New(), Name: "r1", Rule: "/a", RequestsPerWindow: 10, WindowDuration: 30, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), Name: "r2", Rule: "/b", RequestsPerWindow: 20, WindowDuration: 60, Enabled: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListRateLimitsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRateLimitsQuery{
		Filter: domain.RateLimitFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.RateLimits) != 2 {
		t.Fatalf("expected 2 rate limits, got %d", len(result.RateLimits))
	}
	if result.RateLimits[0].Name != "r1" {
		t.Errorf("expected r1, got %s", result.RateLimits[0].Name)
	}
}

func TestListRateLimitsHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{views: []*domain.RateLimitView{}, total: 0}

	handler := NewListRateLimitsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRateLimitsQuery{
		Filter: domain.RateLimitFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.RateLimits) != 0 {
		t.Errorf("expected 0 rate limits, got %d", len(result.RateLimits))
	}
}

func TestListRateLimitsHandler_WithFilters(t *testing.T) {
	t.Parallel()

	enabled := true
	name := "api"
	readRepo := &mockReadRepo{
		views: []*domain.RateLimitView{
			{ID: uuid.New(), Name: "api-rule", Rule: "/api", RequestsPerWindow: 100, WindowDuration: 60, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListRateLimitsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRateLimitsQuery{
		Filter: domain.RateLimitFilter{Name: &name, Enabled: &enabled, Limit: 10},
	})
	require.NoError(t, err)
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListRateLimitsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListRateLimitsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListRateLimitsQuery{Filter: domain.RateLimitFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
