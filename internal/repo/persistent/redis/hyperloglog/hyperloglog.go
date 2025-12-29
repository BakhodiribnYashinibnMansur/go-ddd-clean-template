package hyperloglog

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// HyperLogLog handles Redis HyperLogLog operations for cardinality estimation
type HyperLogLog struct {
	client *redis.Client
}

// New creates a new HyperLogLog instance
func New(client *redis.Client) *HyperLogLog {
	return &HyperLogLog{
		client: client,
	}
}

// PFAdd adds elements to HyperLogLog
func (h *HyperLogLog) PFAdd(ctx context.Context, key string, els ...any) (int64, error) {
	return h.client.PFAdd(ctx, key, els...).Result()
}

// PFCount returns the approximated cardinality of the set(s)
func (h *HyperLogLog) PFCount(ctx context.Context, keys ...string) (int64, error) {
	return h.client.PFCount(ctx, keys...).Result()
}

// PFMerge merges multiple HyperLogLog values into a single one
func (h *HyperLogLog) PFMerge(ctx context.Context, dest string, keys ...string) error {
	return h.client.PFMerge(ctx, dest, keys...).Err()
}

// Delete removes a HyperLogLog key
func (h *HyperLogLog) Delete(ctx context.Context, key string) error {
	return h.client.Del(ctx, key).Err()
}

// Exists checks if a HyperLogLog key exists
func (h *HyperLogLog) Exists(ctx context.Context, key string) (bool, error) {
	count, err := h.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
