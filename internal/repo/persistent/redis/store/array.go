package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ArrayI[T any] interface {
	Get(key string) ([]T, error)
	Set(key string, value []T, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) ([]T, error)
}

type Array[T any] struct {
	db *redis.Client
}

func NewArray[T any](db *redis.Client) *Array[T] {
	return &Array[T]{
		db: db,
	}
}

func (a *Array[T]) Get(key string) ([]T, error) {
	valStrs, err := a.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return a.unmarshalSlice(valStrs)
}

func (a *Array[T]) Set(key string, value []T, expiration time.Duration) error {
	ctx := context.Background()
	if len(value) == 0 {
		return a.db.Del(ctx, key).Err()
	}

	vals := make([]any, len(value))
	for i, v := range value {
		vals[i] = v
	}

	pipe := a.db.TxPipeline()
	pipe.Del(ctx, key)
	pipe.RPush(ctx, key, vals...)
	// Only set expiration if it's greater than 0
	if expiration > 0 {
		pipe.Expire(ctx, key, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (a *Array[T]) Delete(key string) error {
	return a.db.Del(context.Background(), key).Err()
}

func (a *Array[T]) Pop(key string) ([]T, error) {
	ctx := context.Background()
	pipe := a.db.TxPipeline()
	get := pipe.LRange(ctx, key, 0, -1)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	valStrs, err := get.Result()
	if err != nil {
		return nil, err
	}
	return a.unmarshalSlice(valStrs)
}

func (a *Array[T]) unmarshalOne(s string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	return val, err
}

func (a *Array[T]) unmarshalSlice(valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := a.unmarshalOne(s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}
