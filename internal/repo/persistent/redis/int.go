package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Int struct {
	db *redis.Client
}

func NewInt(db *redis.Client) *Int {
	return &Int{
		db: db,
	}
}

func (i *Int) Get(key string) (int64, error) {
	value, err := i.db.Get(context.Background(), key).Int64()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (i *Int) Pop(key string) (int64, error) {
	value, err := i.db.Get(context.Background(), key).Int64()
	if err != nil {
		return 0, err
	}
	err = i.db.Del(context.Background(), key).Err()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (i *Int) Set(key string, value int64, expiration time.Duration) error {
	err := i.db.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (i *Int) Delete(key string) error {
	err := i.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
