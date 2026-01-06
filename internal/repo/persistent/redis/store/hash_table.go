package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type HashTableI[V any] interface {
	Get(ctx context.Context, key string, delete bool) (map[string]V, error)
	Pop(ctx context.Context, key string) (map[string]V, error)
	Set(ctx context.Context, key string, hashKey map[string]V, expirationTime time.Duration) error
	Delete(ctx context.Context, key string) error
}

type HashTable[V any] struct {
	db *redis.Client
}

func NewHashTable[V any](db *redis.Client) *HashTable[V] {
	return &HashTable[V]{
		db: db,
	}
}

func (h *HashTable[V]) Get(ctx context.Context, key string, delete bool) (map[string]V, error) {
	valMap, err := h.db.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash table from key %s: %w", key, err)
	}

	res := make(map[string]V, len(valMap))
	for k, s := range valMap {
		var v V
		cmd := redis.NewStringCmd(ctx)
		cmd.SetVal(s)
		err := cmd.Scan(&v)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hash value for key %s: %w", k, err)
		}
		res[k] = v
	}

	if delete {
		err = h.db.Del(ctx, key).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to delete hash table key %s: %w", key, err)
		}
	}
	return res, nil
}

func (h *HashTable[V]) Pop(ctx context.Context, key string) (map[string]V, error) {
	return h.Get(ctx, key, true)
}

func (h *HashTable[V]) Set(ctx context.Context, key string, hashKey map[string]V, expirationTime time.Duration) error {

	// If empty map, delete the key
	if len(hashKey) == 0 {
		if err := h.db.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete empty hash table key %s: %w", key, err)
		}
		return nil
	}

	marshalledMap := make(map[string]any, len(hashKey))
	for k, v := range hashKey {
		marshalledMap[k] = v
	}

	err := h.db.HMSet(ctx, key, marshalledMap).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash table key %s: %w", key, err)
	}
	// Only set expiration if it's greater than 0
	if expirationTime > 0 {
		err = h.db.Expire(ctx, key, expirationTime).Err()
		if err != nil {
			return fmt.Errorf("failed to set expiration for hash table key %s: %w", key, err)
		}
	}
	return nil
}

func (h *HashTable[V]) Delete(ctx context.Context, key string) error {
	if err := h.db.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete hash table key %s: %w", key, err)
	}
	return nil
}
