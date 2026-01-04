package store

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("failed to get set from key %s: %w", key, err)
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
		return nil, fmt.Errorf("failed to execute pipeline for popping set key %s: %w", key, err)
	}
	valStrs, err := get.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set result for key %s: %w", key, err)
	}
	return s.unmarshalSlice(valStrs)
}

func (s *Set[T]) Set(key string, value []T, expiration time.Duration) error {
	if len(value) == 0 {
		if err := s.db.Del(context.Background(), key).Err(); err != nil {
			return fmt.Errorf("failed to delete empty set key %s: %w", key, err)
		}
		return nil
	}
	marshalled, err := s.marshalSlice(value)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = s.db.SAdd(ctx, key, marshalled...).Err()
	if err != nil {
		return fmt.Errorf("failed to add elements to set key %s: %w", key, err)
	}
	// Only set expiration if it's greater than 0
	if expiration > 0 {
		if err := s.db.Expire(ctx, key, expiration).Err(); err != nil {
			return fmt.Errorf("failed to set expiration for set key %s: %w", key, err)
		}
	}
	return nil
}

func (s *Set[T]) Delete(key string) error {
	if err := s.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete set key %s: %w", key, err)
	}
	return nil
}

func (s *Set[T]) unmarshalOne(vStr string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(vStr)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan set value: %w", err)
	}
	return val, nil
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
