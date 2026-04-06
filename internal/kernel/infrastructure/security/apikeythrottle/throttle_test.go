package apikeythrottle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"gct/internal/kernel/infrastructure/security/apikeythrottle"
)

func setup(t *testing.T) (*miniredis.Miniredis, *apikeythrottle.Throttle) {
	t.Helper()

	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	throttle := apikeythrottle.New(client, apikeythrottle.DefaultConfig())

	return mr, throttle
}

func TestThrottleAfter20Fails(t *testing.T) {
	_, throttle := setup(t)
	ctx := context.Background()
	ip := "10.0.0.1"

	for i := 0; i < 20; i++ {
		if err := throttle.RecordFail(ctx, ip); err != nil {
			t.Fatalf("RecordFail #%d: %v", i+1, err)
		}
	}

	err := throttle.Check(ctx, ip)
	if !errors.Is(err, apikeythrottle.ErrThrottled) {
		t.Fatalf("expected ErrThrottled after 20 fails, got %v", err)
	}
}

func TestBlockExpiresAfterTTL(t *testing.T) {
	mr, throttle := setup(t)
	ctx := context.Background()
	ip := "10.0.0.2"

	for i := 0; i < 20; i++ {
		if err := throttle.RecordFail(ctx, ip); err != nil {
			t.Fatalf("RecordFail: %v", err)
		}
	}

	// Verify blocked.
	if err := throttle.Check(ctx, ip); !errors.Is(err, apikeythrottle.ErrThrottled) {
		t.Fatal("expected blocked")
	}

	// Fast-forward past the block duration.
	mr.FastForward(1*time.Hour + time.Second)

	if err := throttle.Check(ctx, ip); err != nil {
		t.Fatalf("expected unblocked after TTL, got %v", err)
	}
}

func TestNoopThrottleAlwaysPasses(t *testing.T) {
	ctx := context.Background()
	noop := apikeythrottle.NoopThrottle{}

	if err := noop.Check(ctx, "any"); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if err := noop.RecordFail(ctx, "any"); err != nil {
		t.Fatalf("RecordFail: %v", err)
	}
}
