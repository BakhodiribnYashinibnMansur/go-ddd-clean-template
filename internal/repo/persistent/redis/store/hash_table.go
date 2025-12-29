package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type HashTableI[V any] interface {
	Get(key string, delete bool) (map[string]V, error)
	Pop(key string) (map[string]V, error)
	Set(key string, hashKey map[string]V, expirationTime time.Duration) error
	Delete(key string) error
}

type HashTable[V any] struct {
	db *redis.Client
}

func NewHashTable[V any](db *redis.Client) *HashTable[V] {
	return &HashTable[V]{
		db: db,
	}
}

func (h *HashTable[V]) Get(key string, delete bool) (map[string]V, error) {
	ctx := context.Background()
	valMap, err := h.db.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	res := make(map[string]V, len(valMap))
	for k, s := range valMap {
		var v V
		err := redis.NewStringCmd(ctx, s).Scan(&v)
		if err != nil {
			return nil, err
		}
		res[k] = v
	}

	if delete {
		err = h.db.Del(ctx, key).Err()
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (h *HashTable[V]) Pop(key string) (map[string]V, error) {
	return h.Get(key, true)
}

func (h *HashTable[V]) Set(key string, hashKey map[string]V, expirationTime time.Duration) error {
	ctx := context.Background()

	marshalledMap := make(map[string]any, len(hashKey))
	for k, v := range hashKey {
		marshalledMap[k] = v
	}

	err := h.db.HMSet(ctx, key, marshalledMap).Err()
	if err != nil {
		return err
	}
	err = h.db.Expire(ctx, key, expirationTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (h *HashTable[V]) Delete(key string) error {
	return h.db.Del(context.Background(), key).Err()
}
