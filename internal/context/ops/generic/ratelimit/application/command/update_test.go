package command

import (
	"context"
	"testing"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

func TestUpdateRateLimitHandler_Handle(t *testing.T) {
	t.Parallel()

	rl := ratelimitentity.NewRateLimit("old-name", "/old", 50, 30, true)

	repo := &mockRateLimitRepo{
		findFn: func(_ context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
			if id == rl.TypedID() {
				return rl, nil
			}
			return nil, ratelimitentity.ErrRateLimitNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	newName := "new-name"
	newRequests := 200
	newEnabled := false
	cmd := UpdateRateLimitCommand{
		ID:                ratelimitentity.RateLimitID(rl.ID()),
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

	rl := ratelimitentity.NewRateLimit("name", "/rule", 100, 60, true)

	repo := &mockRateLimitRepo{
		findFn: func(_ context.Context, _ ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
			return rl, nil
		},
	}
	handler := NewUpdateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	newWindow := 120
	err := handler.Handle(context.Background(), UpdateRateLimitCommand{
		ID:             ratelimitentity.RateLimitID(rl.ID()),
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
	handler := NewUpdateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRateLimitCommand{ID: ratelimitentity.NewRateLimitID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
