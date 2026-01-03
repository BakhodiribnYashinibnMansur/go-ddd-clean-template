package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type SetI[T any] interface {
	Get(key string) ([]T, error)
	Set(key string, value []T, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) ([]T, error)
}

type Set[T any] struct {
	db *redis.Client
}

func NewSet[T any](db *redis.Client) *Set[T] {
	return &Set[T]{
		db: db,
	}
}

func (s *Set[T]) Get(key string) ([]T, error) {
	valStrs, err := s.db.SMembers(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return s.unmarshalSlice(valStrs)
}

func (s *Set[T]) Pop(key string) ([]T, error) {
	ctx := context.Background()
	pipe := s.db.TxPipeline()
	get := pipe.SMembers(ctx, key)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	valStrs, err := get.Result()
	if err != nil {
		return nil, err
	}
	return s.unmarshalSlice(valStrs)
}

func (s *Set[T]) Set(key string, value []T, expiration time.Duration) error {
	if len(value) == 0 {
		return s.db.Del(context.Background(), key).Err()
	}
	marshalled, err := s.marshalSlice(value)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = s.db.SAdd(ctx, key, marshalled...).Err()
	if err != nil {
		return err
	}
	// Only set expiration if it's greater than 0
	if expiration > 0 {
		return s.db.Expire(ctx, key, expiration).Err()
	}
	return nil
}

func (s *Set[T]) Delete(key string) error {
	return s.db.Del(context.Background(), key).Err()
}

func (s *Set[T]) unmarshalOne(vStr string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(vStr)
	err := cmd.Scan(&val)
	return val, err
}

func (s *Set[T]) unmarshalSlice(valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, vStr := range valStrs {
		v, err := s.unmarshalOne(vStr)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}

func (s *Set[T]) marshalSlice(value []T) ([]any, error) {
	res := make([]any, len(value))
	for i, v := range value {
		res[i] = v
	}
	return res, nil
}
