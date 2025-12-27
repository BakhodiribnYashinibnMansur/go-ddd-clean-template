package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Float struct {
	db *redis.Client
}

func NewFloat(db *redis.Client) *Float {
	return &Float{
		db: db,
	}
}

func (f *Float) Get(key string) (float64, error) {
	value, err := f.db.Get(context.Background(), key).Float64()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (f *Float) Pop(key string) (float64, error) {
	value, err := f.db.Get(context.Background(), key).Float64()
	if err != nil {
		return 0, err
	}
	err = f.db.Del(context.Background(), key).Err()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (f *Float) Set(key string, value float64, expiration time.Duration) error {
	err := f.db.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (f *Float) Delete(key string) error {
	err := f.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
