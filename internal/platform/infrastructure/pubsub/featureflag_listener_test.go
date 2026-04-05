package pubsub_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gct/internal/platform/infrastructure/pubsub"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type mockFlagInvalidator struct {
	callCount atomic.Int32
}

func (m *mockFlagInvalidator) Invalidate() {
	m.callCount.Add(1)
}

// waitForCount polls the atomic counter until it reaches the expected value or timeout.
func waitForCount(counter *atomic.Int32, expected int32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if counter.Load() >= expected {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return counter.Load() >= expected
}

func TestFeatureFlagListener_InvalidatesOnSignal(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	mock := &mockFlagInvalidator{}
	listener := pubsub.NewFeatureFlagListener(client, mock.Invalidate)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go listener.Start(ctx)
	time.Sleep(50 * time.Millisecond) // allow subscription to establish

	client.Publish(ctx, "signal:featureflag.created", "new")

	if !waitForCount(&mock.callCount, 1, time.Second) {
		t.Errorf("expected 1 invalidation call, got %d", mock.callCount.Load())
	}
}

func TestFeatureFlagListener_MatchesAllFlagEvents(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	mock := &mockFlagInvalidator{}
	listener := pubsub.NewFeatureFlagListener(client, mock.Invalidate)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go listener.Start(ctx)
	time.Sleep(50 * time.Millisecond) // allow subscription to establish

	for _, event := range []string{
		"signal:featureflag.created",
		"signal:featureflag.updated",
		"signal:featureflag.deleted",
		"signal:featureflag.toggled",
	} {
		client.Publish(ctx, event, "new")
	}

	if !waitForCount(&mock.callCount, 4, time.Second) {
		t.Errorf("expected 4 invalidation calls, got %d", mock.callCount.Load())
	}
}

func TestFeatureFlagListener_IgnoresNonMatchingChannels(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	mock := &mockFlagInvalidator{}
	listener := pubsub.NewFeatureFlagListener(client, mock.Invalidate)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go listener.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// These should NOT trigger invalidation
	client.Publish(ctx, "signal:errorcode.created", "new")
	client.Publish(ctx, "signal:cache.invalidated", "new")
	// This SHOULD trigger
	client.Publish(ctx, "signal:featureflag.updated", "new")

	if !waitForCount(&mock.callCount, 1, time.Second) {
		t.Errorf("expected 1 invalidation call, got %d", mock.callCount.Load())
	}
	// Give a little extra time to make sure no extra calls arrive
	time.Sleep(50 * time.Millisecond)
	if got := mock.callCount.Load(); got != 1 {
		t.Errorf("expected exactly 1 invalidation call (non-matching channels should be ignored), got %d", got)
	}
}
