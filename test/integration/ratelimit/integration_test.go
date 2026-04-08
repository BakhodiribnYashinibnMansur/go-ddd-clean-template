package ratelimit

import (
	"context"
	"testing"

	"gct/internal/context/ops/generic/ratelimit"
	"gct/internal/context/ops/generic/ratelimit/application/command"
	"gct/internal/context/ops/generic/ratelimit/application/query"
	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
	"gct/internal/kernel/outbox"
)

func newTestBC(t *testing.T) *ratelimit.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return ratelimit.NewBoundedContext(setup.TestPG.Pool, outbox.NewEventCommitter(nil, nil, eb, l), l)
}

func TestIntegration_CreateAndGetRateLimit(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateRateLimit.Handle(ctx, command.CreateRateLimitCommand{
		Name:              "api_global",
		Rule:              "ip",
		RequestsPerWindow: 100,
		WindowDuration:    60,
		Enabled:           true,
	})
	if err != nil {
		t.Fatalf("CreateRateLimit: %v", err)
	}

	result, err := bc.ListRateLimits.Handle(ctx, query.ListRateLimitsQuery{
		Filter: ratelimitrepo.RateLimitFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListRateLimits: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 rate limit, got %d", result.Total)
	}

	rl := result.RateLimits[0]
	if rl.Name != "api_global" {
		t.Errorf("expected name api_global, got %s", rl.Name)
	}
	if rl.Rule != "ip" {
		t.Errorf("expected rule ip, got %s", rl.Rule)
	}

	view, err := bc.GetRateLimit.Handle(ctx, query.GetRateLimitQuery{ID: ratelimitentity.RateLimitID(rl.ID)})
	if err != nil {
		t.Fatalf("GetRateLimit: %v", err)
	}
	if view.ID != rl.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, rl.ID)
	}
}

func TestIntegration_UpdateRateLimit(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateRateLimit.Handle(ctx, command.CreateRateLimitCommand{
		Name:              "login_limit",
		Rule:              "user",
		RequestsPerWindow: 5,
		WindowDuration:    300,
		Enabled:           true,
	})
	if err != nil {
		t.Fatalf("CreateRateLimit: %v", err)
	}

	list, _ := bc.ListRateLimits.Handle(ctx, query.ListRateLimitsQuery{
		Filter: ratelimitrepo.RateLimitFilter{Limit: 10},
	})
	rlID := ratelimitentity.RateLimitID(list.RateLimits[0].ID)

	newName := "login_limit_v2"
	newRequests := 10
	err = bc.UpdateRateLimit.Handle(ctx, command.UpdateRateLimitCommand{
		ID:                rlID,
		Name:              &newName,
		RequestsPerWindow: &newRequests,
	})
	if err != nil {
		t.Fatalf("UpdateRateLimit: %v", err)
	}

	view, _ := bc.GetRateLimit.Handle(ctx, query.GetRateLimitQuery{ID: rlID})
	if view.Name != "login_limit_v2" {
		t.Errorf("name not updated, got %s", view.Name)
	}
	if view.RequestsPerWindow != 10 {
		t.Errorf("requests_per_window not updated, got %d", view.RequestsPerWindow)
	}
}

func TestIntegration_DeleteRateLimit(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateRateLimit.Handle(ctx, command.CreateRateLimitCommand{
		Name:              "to_delete",
		Rule:              "ip",
		RequestsPerWindow: 50,
		WindowDuration:    120,
		Enabled:           false,
	})
	if err != nil {
		t.Fatalf("CreateRateLimit: %v", err)
	}

	list, _ := bc.ListRateLimits.Handle(ctx, query.ListRateLimitsQuery{
		Filter: ratelimitrepo.RateLimitFilter{Limit: 10},
	})
	rlID := ratelimitentity.RateLimitID(list.RateLimits[0].ID)

	err = bc.DeleteRateLimit.Handle(ctx, command.DeleteRateLimitCommand{ID: rlID})
	if err != nil {
		t.Fatalf("DeleteRateLimit: %v", err)
	}

	list2, _ := bc.ListRateLimits.Handle(ctx, query.ListRateLimitsQuery{
		Filter: ratelimitrepo.RateLimitFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 rate limits after delete, got %d", list2.Total)
	}
}
