package store

import (
	"context"
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
		return nil, err
	}
	return l.unmarshalSlice(valStrs)
}

func (l *List[T]) GetFull(key string) (int64, error) {
	return l.db.LLen(context.Background(), key).Result()
}

func (l *List[T]) Pop(key string, limit, offset int64) ([]T, error) {
	ctx := context.Background()
	pipe := l.db.TxPipeline()
	get := pipe.LRange(ctx, key, offset, offset+limit-1)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	valStrs, err := get.Result()
	if err != nil {
		return nil, err
	}
	return l.unmarshalSlice(valStrs)
}

func (l *List[T]) Set(key string, value []T, expiration time.Duration) error {
	ctx := context.Background()
	if len(value) == 0 {
		return l.db.Del(ctx, key).Err()
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
	return err
}

func (l *List[T]) Delete(key string) error {
	return l.db.Del(context.Background(), key).Err()
}

func (l *List[T]) Len(key string) (int64, error) {
	return l.db.LLen(context.Background(), key).Result()
}

func (l *List[T]) unmarshalOne(s string) (T, error) {
	var val T
	// Use go-redis Scan logic equivalent for strings
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	return val, err
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
