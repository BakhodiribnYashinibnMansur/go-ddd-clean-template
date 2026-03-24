package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"gct/internal/repo/persistent/redis"
	pkgcache "gct/internal/shared/infrastructure/cache"
	"gct/internal/shared/infrastructure/logger"
)

// JitterRedisCache, o'zgaruvchan (jitter) muddatli Redis keshi.
type Jitter struct {
	redis  *redis.Repo
	logger logger.Log
}

// NewJitter yangi Jitter yaratadi.
func NewJitter(redis *redis.Repo, logger logger.Log) *Jitter {
	return &Jitter{
		redis:  redis,
		logger: logger,
	}
}

// Set keshga ma'lumot yozadi, muddatga jitter qo'shadi.
// jitterPercent - o'zgaruvchanlik foizi (0.0 dan 1.0 gacha).
func (c *Jitter) Set(ctx context.Context, key string, value any, duration time.Duration, jitterPercent float64) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		c.logger.Errorc(ctx, "failed to marshal data for jitter cache", "error", err)
		return fmt.Errorf("marshal data: %w", err)
	}

	ttl := applyJitter(duration, jitterPercent)

	if err := c.redis.Primitive.Byte.Set(ctx, key, bytes, ttl); err != nil {
		c.logger.Errorc(ctx, "failed to set jitter cache", "key", key, "error", err)
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// Get keshdan ma'lumot o'qiydi.
func (c *Jitter) Get(ctx context.Context, key string, out any) error {
	if out == nil {
		return pkgcache.ErrNilOutput
	}

	bytes, err := c.redis.Primitive.Byte.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("redis get: %w", err)
	}

	if err := json.Unmarshal(bytes, out); err != nil {
		c.logger.Errorc(ctx, "failed to unmarshal data from jitter cache", "key", key, "error", err)
		return fmt.Errorf("unmarshal data: %w", err)
	}

	return nil
}

// Delete keshdan ma'lumot o'chiradi.
func (c *Jitter) Delete(ctx context.Context, key string) error {
	if err := c.redis.Primitive.Byte.Delete(ctx, key); err != nil {
		c.logger.Errorc(ctx, "failed to delete jitter cache", "key", key, "error", err)
		return fmt.Errorf("redis delete: %w", err)
	}
	return nil
}

func applyJitter(duration time.Duration, percent float64) time.Duration {
	if percent <= 0 {
		return duration
	}

	// Jitter miqdorini hisoblash
	variation := time.Duration(float64(duration) * percent)
	if variation == 0 {
		return duration
	}

	// Tasodifiy qo'shimcha vaqt [0, variation]
	extra := time.Duration(rand.Int63n(int64(variation)))
	return duration + extra
}

// SetJitterCache keshga ma'lumot yozadi, muddatga jitter qo'shadi. Wrapper method.
func (c *Cache) SetJitterCache(ctx context.Context, key string, value any, duration time.Duration, jitterPercent float64) error {
	return c.jitter.Set(ctx, key, value, duration, jitterPercent)
}

// GetJitterCache keshdan ma'lumot o'qiydi. Wrapper method.
func (c *Cache) GetJitterCache(ctx context.Context, key string, out any) error {
	return c.jitter.Get(ctx, key, out)
}

// DeleteJitterCache keshdan ma'lumot o'chiradi. Wrapper method.
func (c *Cache) DeleteJitterCache(ctx context.Context, key string) error {
	return c.jitter.Delete(ctx, key)
}
