package pubsub

import (
	"context"
	"fmt"

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
	if err := p.client.Publish(ctx, channel, message).Err(); err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}
	return nil
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
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message from pubsub: %w", err)
	}
	return msg, nil
}

// Unsubscribe unsubscribes from channels
func (p *PubSub) Unsubscribe(ctx context.Context, pubsub *redis.PubSub, channels ...string) error {
	if err := pubsub.Unsubscribe(ctx, channels...); err != nil {
		return fmt.Errorf("failed to unsubscribe from channels %v: %w", channels, err)
	}
	return nil
}

// Close closes the PubSub connection
func (p *PubSub) Close(pubsub *redis.PubSub) error {
	if err := pubsub.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub connection: %w", err)
	}
	return nil
}
