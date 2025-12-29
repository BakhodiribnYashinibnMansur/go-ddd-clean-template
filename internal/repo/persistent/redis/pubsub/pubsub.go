package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// PubSub handles Redis Pub/Sub operations
type PubSub struct {
	client *redis.Client
}

// New creates a new PubSub instance
func New(client *redis.Client) *PubSub {
	return &PubSub{
		client: client,
	}
}

// Publish publishes a message to a channel
func (p *PubSub) Publish(ctx context.Context, channel string, message any) error {
	return p.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels and returns a PubSub instance
func (p *PubSub) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return p.client.Subscribe(ctx, channels...)
}

// PSubscribe subscribes to channels matching patterns
func (p *PubSub) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return p.client.PSubscribe(ctx, patterns...)
}

// ReceiveMessage receives a message from subscribed channels
func (p *PubSub) ReceiveMessage(ctx context.Context, pubsub *redis.PubSub) (*redis.Message, error) {
	return pubsub.ReceiveMessage(ctx)
}

// Unsubscribe unsubscribes from channels
func (p *PubSub) Unsubscribe(ctx context.Context, pubsub *redis.PubSub, channels ...string) error {
	return pubsub.Unsubscribe(ctx, channels...)
}

// Close closes the PubSub connection
func (p *PubSub) Close(pubsub *redis.PubSub) error {
	return pubsub.Close()
}
