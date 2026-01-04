package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type QueueI[T any] interface {
	Len(key string) (int64, error)
	Get(key string, offset, limit int64) ([]T, error)
	GetFull(key string) (int64, error)
	Delete(key string) error
	Pop(key string) error
	PushFront(key string, value []T) error
	PushBack(key string, value []T) error
	PopFront(key string) (T, error)
	PopBack(key string) (T, error)
	DeleteRange(key string, offset, limit int64) error
	Peek(key string) (T, error)
	IsEmpty(key string) (bool, error)
	Contains(key string, value T) (bool, error)
	ToArray(key string) ([]T, error)
}

type Queue[T any] struct {
	db *redis.Client
}

func NewQueue[T any](db *redis.Client) *Queue[T] {
	return &Queue[T]{
		db: db,
	}
}

func (q *Queue[T]) Len(key string) (int64, error) {
	result, err := q.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length from key %s: %w", key, err)
	}
	return result, nil
}

func (q *Queue[T]) Get(key string, offset, limit int64) ([]T, error) {
	valStrs, err := q.db.LRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue from key %s: %w", key, err)
	}
	return q.unmarshalSlice(valStrs)
}

func (q *Queue[T]) GetFull(key string) (int64, error) {
	result, err := q.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get full queue length from key %s: %w", key, err)
	}
	return result, nil
}

func (q *Queue[T]) Delete(key string) error {
	if err := q.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete queue key %s: %w", key, err)
	}
	return nil
}

func (q *Queue[T]) Pop(key string) error {
	if err := q.db.RPop(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to pop from queue key %s: %w", key, err)
	}
	return nil
}

func (q *Queue[T]) PushFront(key string, value []T) error {
	if len(value) == 0 {
		return nil
	}
	marshalled, err := q.marshalSlice(value)
	if err != nil {
		return err
	}
	if err := q.db.LPush(context.Background(), key, marshalled...).Err(); err != nil {
		return fmt.Errorf("failed to push front to queue key %s: %w", key, err)
	}
	return nil
}

func (q *Queue[T]) PushBack(key string, value []T) error {
	if len(value) == 0 {
		return nil
	}
	marshalled, err := q.marshalSlice(value)
	if err != nil {
		return err
	}
	if err := q.db.RPush(context.Background(), key, marshalled...).Err(); err != nil {
		return fmt.Errorf("failed to push back to queue key %s: %w", key, err)
	}
	return nil
}

func (q *Queue[T]) PopFront(key string) (T, error) {
	var val T
	valStr, err := q.db.LPop(context.Background(), key).Result()
	if err != nil {
		return val, fmt.Errorf("failed to pop front from queue key %s: %w", key, err)
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) PopBack(key string) (T, error) {
	var val T
	valStr, err := q.db.RPop(context.Background(), key).Result()
	if err != nil {
		return val, fmt.Errorf("failed to pop back from queue key %s: %w", key, err)
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) DeleteRange(key string, offset, limit int64) error {
	if err := q.db.LTrim(context.Background(), key, offset, limit).Err(); err != nil {
		return fmt.Errorf("failed to delete range from queue key %s: %w", key, err)
	}
	return nil
}

func (q *Queue[T]) Peek(key string) (T, error) {
	var val T
	valStr, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return val, fmt.Errorf("failed to peek queue key %s: %w", key, err)
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) IsEmpty(key string) (bool, error) {
	l, err := q.db.LLen(context.Background(), key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if queue key %s is empty: %w", key, err)
	}
	return l == 0, nil
}

func (q *Queue[T]) Contains(key string, value T) (bool, error) {
	// Redis doesn't have a direct "LCONTAINS".
	// The original implementation was just checking LIndex 0?
	valStr, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if queue key %s contains value: %w", key, err)
	}

	// Compare string representation
	return fmt.Sprint(value) == valStr, nil
}

func (q *Queue[T]) ToArray(key string) ([]T, error) {
	valStrs, err := q.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to convert queue key %s to array: %w", key, err)
	}
	return q.unmarshalSlice(valStrs)
}

func (q *Queue[T]) unmarshalOne(s string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan queue value: %w", err)
	}
	return val, nil
}

func (q *Queue[T]) unmarshalSlice(valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := q.unmarshalOne(s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}

func (q *Queue[T]) marshalSlice(value []T) ([]any, error) {
	res := make([]any, len(value))
	for i, v := range value {
		res[i] = v
	}
	return res, nil
}
