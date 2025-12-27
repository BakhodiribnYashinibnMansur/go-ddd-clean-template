package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type HashTable struct {
	db *redis.Client
}

func NewHashTable(db *redis.Client) *HashTable {
	return &HashTable{
		db: db,
	}
}

func (h *HashTable) Get(key string, delete bool) (map[string]string, error) {
	hashKey, err := h.db.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return hashKey, nil
}
func (h *HashTable) Pop(key string) (map[string]string, error) {
	hashKey, err := h.db.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	err = h.db.Del(context.Background(), key).Err()
	if err != nil {
		return nil, err
	}
	return hashKey, nil
}

func (h *HashTable) Set(key string, hashKey map[string]any, expirationTime time.Duration) error {
	err := h.db.HMSet(context.Background(), key, hashKey).Err()
	if err != nil {
		return err
	}
	err = h.db.Expire(context.Background(), key, expirationTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (h *HashTable) Delete(key string) error {
	err := h.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
