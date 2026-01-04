package store

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("failed to get priority queue from key %s: %w", key, err)
	}
	return p.unmarshalSlice(valStrs)
}

func (p *PriorityQueue[T]) GetFull(key string) (int64, error) {
	result, err := p.db.ZCard(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get priority queue size from key %s: %w", key, err)
	}
	return result, nil
}

func (p *PriorityQueue[T]) Delete(key string) error {
	if err := p.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) Push(key string, value []GenericZ[T]) error {
	if len(value) == 0 {
		return nil
	}
	zSlice := make([]redis.Z, len(value))
	for i, v := range value {
		zSlice[i] = redis.Z{
			Score:  v.Score,
			Member: v.Member,
		}
	}
	if err := p.db.ZAdd(context.Background(), key, zSlice...).Err(); err != nil {
		return fmt.Errorf("failed to push to priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) PopMin(key string) error {
	if err := p.db.ZPopMin(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to pop min from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) PopMax(key string) error {
	if err := p.db.ZPopMax(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to pop max from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) DeleteRange(key string, offset, limit int64) error {
	if err := p.db.ZRemRangeByRank(context.Background(), key, offset, limit).Err(); err != nil {
		return fmt.Errorf("failed to delete range from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) Peek(key string) (T, error) {
	var val T
	valStrs, err := p.db.ZRange(context.Background(), key, 0, 0).Result()
	if err != nil {
		return val, fmt.Errorf("failed to peek priority queue key %s: %w", key, err)
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
	result, err := p.db.ZCard(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get size of priority queue key %s: %w", key, err)
	}
	return result, nil
}

func (p *PriorityQueue[T]) Clear(key string) error {
	if err := p.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to clear priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) ChangePriority(key string, value T, priority int64) error {
	if err := p.db.ZAdd(context.Background(), key, redis.Z{Score: float64(priority), Member: value}).Err(); err != nil {
		return fmt.Errorf("failed to change priority for key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) ToArray(key string) ([]T, error) {
	valStrs, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to convert priority queue to array for key %s: %w", key, err)
	}
	return p.unmarshalSlice(valStrs)
}

func (p *PriorityQueue[T]) unmarshalOne(s string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(context.Background())
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan priority queue value: %w", err)
	}
	return val, nil
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
