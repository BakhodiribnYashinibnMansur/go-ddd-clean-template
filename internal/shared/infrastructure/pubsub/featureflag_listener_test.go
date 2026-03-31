package pubsub_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/pubsub"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type mockFlagInvalidator struct {
	callCount atomic.Int32
}

func (m *mockFlagInvalidator) Invalidate() {
	m.callCount.Add(1)
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

	time.Sleep(50 * time.Millisecond)
	client.Publish(ctx, "signal:featureflags", "new")
	time.Sleep(100 * time.Millisecond)

	if mock.callCount.Load() != 1 {
		t.Errorf("expected 1 invalidation call, got %d", mock.callCount.Load())
	}
}
