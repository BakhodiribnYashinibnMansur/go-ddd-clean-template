package store

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type GenericZ[T any] struct {
	Score  float64
	Member T
}

type PriorityQueueI[T any] interface {
	Get(key string, offset, limit int64) ([]T, error)
	GetFull(key string) (int64, error)
	Delete(key string) error
	Push(key string, value []GenericZ[T]) error
	PopMin(key string) error
	PopMax(key string) error
	DeleteRange(key string, offset, limit int64) error
	Peek(key string) (T, error)
	IsEmpty(key string) (bool, error)
	Size(key string) (int64, error)
	Clear(key string) error
	ChangePriority(key string, value T, priority int64) error
	ToArray(key string) ([]T, error)
}

type PriorityQueue[T any] struct {
	db *redis.Client
}

func NewPriorityQueue[T any](db *redis.Client) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		db: db,
	}
}

func (p *PriorityQueue[T]) Get(key string, offset, limit int64) ([]T, error) {
	valStrs, err := p.db.ZRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, err
	}
	return p.unmarshalSlice(valStrs)
}

func (p *PriorityQueue[T]) GetFull(key string) (int64, error) {
	return p.db.ZCard(context.Background(), key).Result()
}

func (p *PriorityQueue[T]) Delete(key string) error {
	return p.db.Del(context.Background(), key).Err()
}

func (p *PriorityQueue[T]) Push(key string, value []GenericZ[T]) error {
	zSlice := make([]redis.Z, len(value))
	for i, v := range value {
		zSlice[i] = redis.Z{
			Score:  v.Score,
			Member: v.Member,
		}
	}
	return p.db.ZAdd(context.Background(), key, zSlice...).Err()
}

func (p *PriorityQueue[T]) PopMin(key string) error {
	return p.db.ZPopMin(context.Background(), key).Err()
}

func (p *PriorityQueue[T]) PopMax(key string) error {
	return p.db.ZPopMax(context.Background(), key).Err()
}

func (p *PriorityQueue[T]) DeleteRange(key string, offset, limit int64) error {
	return p.db.ZRemRangeByRank(context.Background(), key, offset, limit).Err()
}

func (p *PriorityQueue[T]) Peek(key string) (T, error) {
	var val T
	valStrs, err := p.db.ZRange(context.Background(), key, 0, 0).Result()
	if err != nil {
		return val, err
	}
	if len(valStrs) == 0 {
		return val, nil
	}
	return p.unmarshalOne(valStrs[0])
}

func (p *PriorityQueue[T]) IsEmpty(key string) (bool, error) {
	sz, err := p.Size(key)
	return sz == 0, err
}

func (p *PriorityQueue[T]) Size(key string) (int64, error) {
	return p.db.ZCard(context.Background(), key).Result()
}

func (p *PriorityQueue[T]) Clear(key string) error {
	return p.db.Del(context.Background(), key).Err()
}

func (p *PriorityQueue[T]) ChangePriority(key string, value T, priority int64) error {
	return p.db.ZAdd(context.Background(), key, redis.Z{Score: float64(priority), Member: value}).Err()
}

func (p *PriorityQueue[T]) ToArray(key string) ([]T, error) {
	valStrs, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return p.unmarshalSlice(valStrs)
}

func (p *PriorityQueue[T]) unmarshalOne(s string) (T, error) {
	var val T
	err := redis.NewStringCmd(context.Background(), s).Scan(&val)
	return val, err
}

func (p *PriorityQueue[T]) unmarshalSlice(valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := p.unmarshalOne(s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}
