package featureflag

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisRetriever retrieves feature flag configuration from Redis.
type RedisRetriever struct {
	client *redis.Client
	key    string
}

// NewRedisRetriever creates a new Redis retriever.
func NewRedisRetriever(client *redis.Client, key string) *RedisRetriever {
	return &RedisRetriever{
		client: client,
		key:    key,
	}
}

// Retrieve fetches the feature flag configuration from Redis.
func (r *RedisRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("feature flag configuration not found in redis: %w", err)
		}
		return nil, fmt.Errorf("failed to retrieve feature flag configuration from redis: %w", err)
	}

	return data, nil
}
