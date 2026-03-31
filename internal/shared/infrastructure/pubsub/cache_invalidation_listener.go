package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// CacheInvalidationListener subscribes to signal:cache:invalidate and deletes
// the specified cache key when a signal arrives from another instance.
type CacheInvalidationListener struct {
	client    *redis.Client
	deleteKey func(key string)
}

// NewCacheInvalidationListener creates a new cache invalidation listener.
func NewCacheInvalidationListener(client *redis.Client, deleteKey func(key string)) *CacheInvalidationListener {
	return &CacheInvalidationListener{
		client:    client,
		deleteKey: deleteKey,
	}
}

// Start begins listening for cache invalidation signals. Blocks until ctx is cancelled.
func (l *CacheInvalidationListener) Start(ctx context.Context) {
	sub := NewSubscriber(l.client)
	sub.Subscribe(ctx, "signal:cache:invalidate", func(channel, payload string) {
		l.deleteKey(payload)
	})
}
