package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Bool struct {
	db *redis.Client
}

func NewBool(db *redis.Client) *Bool {
	return &Bool{
		db: db,
	}
}

// Set sets a boolean value in Redis
func (b *Bool) Set(key string, value bool, expiration time.Duration) error {
	var val string = "0"
	if value {
		val = "1"
	}
	err := b.db.Set(context.Background(), key, val, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get gets a boolean value from Redis
func (b *Bool) Get(key string) (bool, error) {
	val, err := b.db.Get(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// GetWithDefault gets a boolean value from Redis with a default value if key doesn't exist
func (b *Bool) GetWithDefault(key string, defaultValue bool) (bool, error) {
	val, err := b.db.Get(context.Background(), key).Result()
	if err != nil {
		return defaultValue, nil
	}
	return val == "1", nil
}

// Delete deletes a boolean value from Redis
func (b *Bool) Delete(key string) error {
	err := b.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
