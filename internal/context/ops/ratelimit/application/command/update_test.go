package command

import (
	"context"
	"testing"

	"gct/internal/context/ops/ratelimit/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateRateLimitHandler_Handle(t *testing.T) {
	t.Parallel()

	rl := domain.NewRateLimit("old-name", "/old", 50, 30, true)

	repo := &mockRateLimitRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.RateLimit, error) {
			if id == rl.ID() {
				return rl, nil
			}
			return nil, domain.ErrRateLimitNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateRateLimitHandler(repo, eb, &mockLogger{})

	newName := "new-name"
	newRequests := 200
	newEnabled := false
	cmd := UpdateRateLimitCommand{
		ID:                domain.RateLimitID(rl.ID()),
		Name:              &newName,
		RequestsPerWindow: &newRequests,
		Enabled:           &newEnabled,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updated == nil {
		t.Fatal("expected rate limit to be updated")
	}
	if repo.updated.Name() != "new-name" {
		t.Errorf("expected name new-name, got %s", repo.updated.Name())
	}
	if repo.updated.RequestsPerWindow() != 200 {
		t.Errorf("expected requestsPerWindow 200, got %d", repo.updated.RequestsPerWindow())
	}
	if repo.updated.Enabled() {
		t.Error("expected enabled false")
	}
	// Rule should remain unchanged
	if repo.updated.Rule() != "/old" {
		t.Errorf("expected rule /old (unchanged), got %s", repo.updated.Rule())
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "ratelimit.changed" {
		t.Errorf("expected ratelimit.changed, got %s", eb.published[0].EventName())
	}
}

func TestUpdateRateLimitHandler_PartialUpdate(t *testing.T) {
	t.Parallel()

	rl := domain.NewRateLimit("name", "/rule", 100, 60, true)

	repo := &mockRateLimitRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.RateLimit, error) {
			return rl, nil
		},
	}
	handler := NewUpdateRateLimitHandler(repo, &mockEventBus{}, &mockLogger{})

	newWindow := 120
	err := handler.Handle(context.Background(), UpdateRateLimitCommand{
		ID:             domain.RateLimitID(rl.ID()),
		WindowDuration: &newWindow,
	})
	require.NoError(t, err)
	if repo.updated.WindowDuration() != 120 {
		t.Errorf("expected windowDuration 120, got %d", repo.updated.WindowDuration())
	}
	if repo.updated.Name() != "name" {
		t.Error("name should remain unchanged")
	}
}

func TestUpdateRateLimitHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRateLimitRepo{}
	handler := NewUpdateRateLimitHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRateLimitCommand{ID: domain.NewRateLimitID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
