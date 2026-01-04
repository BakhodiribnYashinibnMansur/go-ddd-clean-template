package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gct/internal/domain"
	pkgCache "gct/pkg/cache"

	"go.uber.org/zap"
)

// CreatePublicCache stores data in Redis cache with pagination info
func (c *Cache) CreatePublicCache(
	data any,
	key string,
	lang string,
	pagination *domain.Pagination,
	duration time.Duration,
) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	cacheKey := createCacheKey(key, lang, pagination)
	if err := c.redis.Primitive.Byte.Set(cacheKey, bytes, duration); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// GetPublicCache retrieves data from Redis cache using pagination info
func (c *Cache) GetPublicCache(
	key string,
	lang string,
	pagination *domain.Pagination,
	out any,
) error {
	if out == nil {
		return pkgCache.ErrNilOutput
	}

	cacheKey := createCacheKey(key, lang, pagination)

	bytes, err := c.redis.Primitive.Byte.Get(cacheKey)
	if err != nil {
		return fmt.Errorf("redis get: %w", err)
	}

	if err := json.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	return nil
}

// DeletePublicCache removes an item from Redis cache
func (c *Cache) DeletePublicCache(key string, lang string, pagination *domain.Pagination) error {
	cacheKey := createCacheKey(key, lang, pagination)
	if err := c.redis.Primitive.Byte.Delete(cacheKey); err != nil {
		c.logger.Errorw("failed to delete public cache", zap.Error(err))
		return fmt.Errorf("redis delete: %w", err)
	}
	return nil
}

// DeletePublicCaches removes all items matching the key pattern from Redis cache.
// When triggered by Postgres notification, 'key' corresponds to the table name.
// It can also be used manually with any specific key prefix.
func (c *Cache) DeletePublicCaches(key string) error {
	keys, err := c.redis.Primitive.String.Scan(key + "*")
	if err != nil {
		c.logger.Errorw("failed to scan public cache keys", zap.Error(err))
		return fmt.Errorf("redis scan: %w", err)
	}

	var firstErr error
	for _, key := range keys {
		if err := c.redis.Primitive.Byte.Delete(key); err != nil {
			c.logger.Errorw("failed to delete public cache item", "key", key, zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if firstErr != nil {
		return fmt.Errorf("redis delete items: %w", firstErr)
	}

	return nil
}

// createCacheKey generates a unique cache key combining the base key and pagination info
func createCacheKey(key string, lang string, pagination *domain.Pagination) string {
	if lang != "" {
		key = key + "_" + lang
	}
	if pagination != nil {
		key = key + "_" + strconv.FormatInt(pagination.Offset, 10) + "_" + strconv.FormatInt(pagination.Limit, 10)
	}
	return key
}
