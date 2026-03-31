package eventbus_test

import (
	"context"
	"encoding/json"
	"testing"

	"gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/eventbus"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRedisClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestRedisStreamsEventBus_Publish(t *testing.T) {
	client, _ := setupRedisClient(t)
	defer client.Close()

	bus := eventbus.NewRedisStreamsEventBus(client, 1000)

	evt := newTestEvent("notification.sent")
	err := bus.Publish(context.Background(), evt)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	// Verify stream entry was created
	msgs, err := client.XRange(context.Background(), "stream:notification.sent", "-", "+").Result()
	if err != nil {
		t.Fatalf("xrange failed: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 stream message, got %d", len(msgs))
	}

	// Verify payload
	data, ok := msgs[0].Values["data"]
	if !ok {
		t.Fatal("expected 'data' field in stream message")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(data.(string)), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["event_name"] != "notification.sent" {
		t.Errorf("expected event_name 'notification.sent', got %v", payload["event_name"])
	}
}

func TestRedisStreamsEventBus_LocalHandlers(t *testing.T) {
	client, _ := setupRedisClient(t)
	defer client.Close()

	bus := eventbus.NewRedisStreamsEventBus(client, 1000)

	var received string
	_ = bus.Subscribe("order.placed", func(ctx context.Context, event domain.DomainEvent) error {
		received = event.EventName()
		return nil
	})

	evt := newTestEvent("order.placed")
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if received != "order.placed" {
		t.Errorf("expected local handler to receive 'order.placed', got %q", received)
	}
}

func TestRedisStreamsEventBus_ReadStream(t *testing.T) {
	client, _ := setupRedisClient(t)
	defer client.Close()

	bus := eventbus.NewRedisStreamsEventBus(client, 1000)

	// Publish 3 events
	for i := 0; i < 3; i++ {
		_ = bus.Publish(context.Background(), newTestEvent("test.event"))
	}

	// Read all from beginning
	msgs, err := bus.ReadStream(context.Background(), "stream:test.event", "0", 10)
	if err != nil {
		t.Fatalf("read stream failed: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages, got %d", len(msgs))
	}
}
