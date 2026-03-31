package pubsub_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/pubsub"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type mockCache struct {
	mu      sync.Mutex
	deleted []string
}

func (m *mockCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted = append(m.deleted, key)
}

func TestCacheInvalidationListener_DeletesOnSignal(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	mc := &mockCache{}
	listener := pubsub.NewCacheInvalidationListener(client, mc.Delete)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go listener.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	client.Publish(ctx, "signal:cache:invalidate", "users:abc-123")
	time.Sleep(100 * time.Millisecond)

	mc.mu.Lock()
	defer mc.mu.Unlock()
	if len(mc.deleted) != 1 {
		t.Fatalf("expected 1 delete, got %d", len(mc.deleted))
	}
	if mc.deleted[0] != "users:abc-123" {
		t.Errorf("expected key 'users:abc-123', got %q", mc.deleted[0])
	}
}
