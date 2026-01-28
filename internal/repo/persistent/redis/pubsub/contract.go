package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// PubSubI defines Redis Pub/Sub operations interface
type PubSubI interface {
	Publish(ctx context.Context, channel string, message any) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub
	ReceiveMessage(ctx context.Context, pubsub *redis.PubSub) (*redis.Message, error)
	Unsubscribe(ctx context.Context, pubsub *redis.PubSub, channels ...string) error
	Close(pubsub *redis.PubSub) error
}
