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
	return q.db.LLen(context.Background(), key).Result()
}

func (q *Queue[T]) Get(key string, offset, limit int64) ([]T, error) {
	valStrs, err := q.db.LRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, err
	}
	return q.unmarshalSlice(valStrs)
}

func (q *Queue[T]) GetFull(key string) (int64, error) {
	return q.db.LLen(context.Background(), key).Result()
}

func (q *Queue[T]) Delete(key string) error {
	return q.db.Del(context.Background(), key).Err()
}

func (q *Queue[T]) Pop(key string) error {
	return q.db.RPop(context.Background(), key).Err()
}

func (q *Queue[T]) PushFront(key string, value []T) error {
	if len(value) == 0 {
		return nil
	}
	marshalled, err := q.marshalSlice(value)
	if err != nil {
		return err
	}
	return q.db.LPush(context.Background(), key, marshalled...).Err()
}

func (q *Queue[T]) PushBack(key string, value []T) error {
	if len(value) == 0 {
		return nil
	}
	marshalled, err := q.marshalSlice(value)
	if err != nil {
		return err
	}
	return q.db.RPush(context.Background(), key, marshalled...).Err()
}

func (q *Queue[T]) PopFront(key string) (T, error) {
	var val T
	valStr, err := q.db.LPop(context.Background(), key).Result()
	if err != nil {
		return val, err
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) PopBack(key string) (T, error) {
	var val T
	valStr, err := q.db.RPop(context.Background(), key).Result()
	if err != nil {
		return val, err
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) DeleteRange(key string, offset, limit int64) error {
	return q.db.LTrim(context.Background(), key, offset, limit).Err()
}

func (q *Queue[T]) Peek(key string) (T, error) {
	var val T
	valStr, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return val, err
	}
	return q.unmarshalOne(valStr)
}

func (q *Queue[T]) IsEmpty(key string) (bool, error) {
	l, err := q.db.LLen(context.Background(), key).Result()
	return l == 0, err
}

func (q *Queue[T]) Contains(key string, value T) (bool, error) {
	// Redis doesn't have a direct "LCONTAINS".
	// The original implementation was just checking LIndex 0?
	valStr, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return false, err
	}

	// Compare string representation
	return fmt.Sprint(value) == valStr, nil
}

func (q *Queue[T]) ToArray(key string) ([]T, error) {
	valStrs, err := q.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return q.unmarshalSlice(valStrs)
}

func (q *Queue[T]) unmarshalOne(s string) (T, error) {
	var val T
	err := redis.NewStringCmd(context.Background(), s).Scan(&val)
	return val, err
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
