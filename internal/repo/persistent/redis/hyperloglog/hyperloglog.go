package hyperloglog

import (
	"context"
	"fmt"

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
	result, err := h.client.PFAdd(ctx, key, els...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add elements to HyperLogLog key %s: %w", key, err)
	}
	return result, nil
}

// PFCount returns the approximated cardinality of the set(s)
func (h *HyperLogLog) PFCount(ctx context.Context, keys ...string) (int64, error) {
	result, err := h.client.PFCount(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count HyperLogLog keys %v: %w", keys, err)
	}
	return result, nil
}

// PFMerge merges multiple HyperLogLog values into a single one
func (h *HyperLogLog) PFMerge(ctx context.Context, dest string, keys ...string) error {
	if err := h.client.PFMerge(ctx, dest, keys...).Err(); err != nil {
		return fmt.Errorf("failed to merge HyperLogLog keys %v into %s: %w", keys, dest, err)
	}
	return nil
}

// Delete removes a HyperLogLog key
func (h *HyperLogLog) Delete(ctx context.Context, key string) error {
	if err := h.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete HyperLogLog key %s: %w", key, err)
	}
	return nil
}

// Exists checks if a HyperLogLog key exists
func (h *HyperLogLog) Exists(ctx context.Context, key string) (bool, error) {
	count, err := h.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of HyperLogLog key %s: %w", key, err)
	}
	return count > 0, nil
}
