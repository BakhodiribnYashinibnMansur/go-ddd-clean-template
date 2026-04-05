package sse_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/sse"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestBridge_ForwardsStreamToHub(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	hub := sse.NewHub(256)
	bridge := sse.NewBridge(client, hub, logger.Noop())

	ch := hub.Register("audit")
	defer hub.Unregister("audit", ch)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go bridge.Listen(ctx, "audit", "signal:audit", "stream:audit")

	time.Sleep(50 * time.Millisecond)

	// Simulate what RedisStreamsEventBus.Publish does
	payload, _ := json.Marshal(map[string]any{"action": "USER_DELETE"})
	client.XAdd(ctx, &redis.XAddArgs{
		Stream: "stream:audit",
		Values: map[string]any{"data": string(payload)},
	})
	client.Publish(ctx, "signal:audit", "new")

	select {
	case msg := <-ch:
		if msg.Event != "audit" {
			t.Errorf("expected event 'audit', got %q", msg.Event)
		}
		if msg.ID == "" {
			t.Error("expected non-empty stream ID")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}
