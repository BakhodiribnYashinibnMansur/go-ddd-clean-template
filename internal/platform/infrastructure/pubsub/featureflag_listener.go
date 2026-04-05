package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// FeatureFlagListener subscribes to signal:featureflags and calls
// the invalidate function when a flag changes on any instance.
type FeatureFlagListener struct {
	client     *redis.Client
	invalidate func()
}

// NewFeatureFlagListener creates a new feature flag sync listener.
func NewFeatureFlagListener(client *redis.Client, invalidate func()) *FeatureFlagListener {
	return &FeatureFlagListener{
		client:     client,
		invalidate: invalidate,
	}
}

// Start begins listening for feature flag change signals. Blocks until ctx is cancelled.
// Uses pattern subscribe to match all signal:featureflag.* channels
// (e.g. signal:featureflag.created, signal:featureflag.updated, etc.).
func (l *FeatureFlagListener) Start(ctx context.Context) {
	sub := NewSubscriber(l.client)
	sub.PSubscribe(ctx, "signal:featureflag.*", func(channel, payload string) {
		l.invalidate()
	})
}
