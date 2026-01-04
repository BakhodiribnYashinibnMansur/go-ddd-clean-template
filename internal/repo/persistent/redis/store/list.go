package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ListI[T any] interface {
	Get(key string) ([]T, error)
	Set(key string, value []T, expiration time.Duration) error
	Delete(key string) error
	Pop(key string, limit, offset int64) ([]T, error)
	GetFull(key string) (int64, error)
	Len(key string) (int64, error)
}

type List[T any] struct {
	db *redis.Client
}

func NewList[T any](db *redis.Client) *List[T] {
	return &List[T]{
		db: db,
	}
}

func (l *List[T]) Get(key string) ([]T, error) {
	valStrs, err := l.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list from key %s: %w", key, err)
	}
	return l.unmarshalSlice(valStrs)
}

func (l *List[T]) GetFull(key string) (int64, error) {
	result, err := l.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length from key %s: %w", key, err)
	}
	return result, nil
}

func (l *List[T]) Pop(key string, limit, offset int64) ([]T, error) {
	ctx := context.Background()
	pipe := l.db.TxPipeline()
	get := pipe.LRange(ctx, key, offset, offset+limit-1)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline for popping list key %s: %w", key, err)
	}
	valStrs, err := get.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list result for key %s: %w", key, err)
	}
	return l.unmarshalSlice(valStrs)
}

func (l *List[T]) Set(key string, value []T, expiration time.Duration) error {
	ctx := context.Background()
	if len(value) == 0 {
		if err := l.db.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete empty list key %s: %w", key, err)
		}
		return nil
	}

	vals := make([]any, len(value))
	for i, v := range value {
		vals[i] = v
	}

	pipe := l.db.TxPipeline()
	pipe.Del(ctx, key)
	// Use RPush instead of LPush to preserve order
	pipe.RPush(ctx, key, vals...)
	// Only set expiration if it's greater than 0
	if expiration > 0 {
		pipe.Expire(ctx, key, expiration)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline for list key %s: %w", key, err)
	}
	return nil
}

func (l *List[T]) Delete(key string) error {
	if err := l.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete list key %s: %w", key, err)
	}
	return nil
}

func (l *List[T]) Len(key string) (int64, error) {
	result, err := l.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length from key %s: %w", key, err)
	}
	return result, nil
}

func (l *List[T]) unmarshalOne(s string) (T, error) {
	var val T
	// Use go-redis Scan logic equivalent for strings
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan list value: %w", err)
	}
	return val, nil
}

func (l *List[T]) unmarshalSlice(valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := l.unmarshalOne(s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}
