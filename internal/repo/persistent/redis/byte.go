package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Byte struct {
	db *redis.Client
}

func NewByte(db *redis.Client) *Byte {
	return &Byte{
		db: db,
	}
}

func (b *Byte) Get(key string) ([]byte, error) {
	value, err := b.db.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func (b *Byte) Set(key string, value []byte, expiration time.Duration) error {
	err := b.db.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (b *Byte) Delete(key string) error {
	err := b.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
