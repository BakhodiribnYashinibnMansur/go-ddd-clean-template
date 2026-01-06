package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ArrayI[T any] interface {
	Get(ctx context.Context, key string) ([]T, error)
	Set(ctx context.Context, key string, value []T, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Pop(ctx context.Context, key string) ([]T, error)
}

type Array[T any] struct {
	db *redis.Client
}

func NewArray[T any](db *redis.Client) *Array[T] {
	return &Array[T]{
		db: db,
	}
}

func (a *Array[T]) Get(ctx context.Context, key string) ([]T, error) {
	valStrs, err := a.db.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get array from key %s: %w", key, err)
	}
	return a.unmarshalSlice(ctx, valStrs)
}

func (a *Array[T]) Set(ctx context.Context, key string, value []T, expiration time.Duration) error {
	if len(value) == 0 {
		if err := a.db.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete empty array key %s: %w", key, err)
		}
		return nil
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
	if err != nil {
		return fmt.Errorf("failed to execute pipeline for array key %s: %w", key, err)
	}
	return nil
}

func (a *Array[T]) Delete(ctx context.Context, key string) error {
	if err := a.db.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete array key %s: %w", key, err)
	}
	return nil
}

func (a *Array[T]) Pop(ctx context.Context, key string) ([]T, error) {
	pipe := a.db.TxPipeline()
	get := pipe.LRange(ctx, key, 0, -1)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline for popping array key %s: %w", key, err)
	}

	valStrs, err := get.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get array result for key %s: %w", key, err)
	}
	return a.unmarshalSlice(ctx, valStrs)
}

func (a *Array[T]) unmarshalOne(ctx context.Context, s string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(ctx)
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan array value: %w", err)
	}
	return val, nil
}

func (a *Array[T]) unmarshalSlice(ctx context.Context, valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := a.unmarshalOne(ctx, s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}
