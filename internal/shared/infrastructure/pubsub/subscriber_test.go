package pubsub_test

import (
	"context"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/pubsub"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestSubscriber_ReceivesMessage(t *testing.T) {
	client, _ := setupRedis(t)
	defer client.Close()

	sub := pubsub.NewSubscriber(client)

	received := make(chan string, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go sub.Subscribe(ctx, "test:channel", func(channel, payload string) {
		received <- payload
	})

	time.Sleep(50 * time.Millisecond)
	client.Publish(ctx, "test:channel", "hello")

	select {
	case msg := <-received:
		if msg != "hello" {
			t.Errorf("expected 'hello', got %q", msg)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}
