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
	Get(ctx context.Context, key string, offset, limit int64) ([]T, error)
	GetFull(ctx context.Context, key string) (int64, error)
	Delete(ctx context.Context, key string) error
	Push(ctx context.Context, key string, value []GenericZ[T]) error
	PopMin(ctx context.Context, key string) error
	PopMax(ctx context.Context, key string) error
	DeleteRange(ctx context.Context, key string, offset, limit int64) error
	Peek(ctx context.Context, key string) (T, error)
	IsEmpty(ctx context.Context, key string) (bool, error)
	Size(ctx context.Context, key string) (int64, error)
	Clear(ctx context.Context, key string) error
	ChangePriority(ctx context.Context, key string, value T, priority int64) error
	ToArray(ctx context.Context, key string) ([]T, error)
}

type PriorityQueue[T any] struct {
	db *redis.Client
}

func NewPriorityQueue[T any](db *redis.Client) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		db: db,
	}
}

func (p *PriorityQueue[T]) Get(ctx context.Context, key string, offset, limit int64) ([]T, error) {
	valStrs, err := p.db.ZRange(ctx, key, offset, limit).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get priority queue from key %s: %w", key, err)
	}
	return p.unmarshalSlice(ctx, valStrs)
}

func (p *PriorityQueue[T]) GetFull(ctx context.Context, key string) (int64, error) {
	result, err := p.db.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get priority queue size from key %s: %w", key, err)
	}
	return result, nil
}

func (p *PriorityQueue[T]) Delete(ctx context.Context, key string) error {
	if err := p.db.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) Push(ctx context.Context, key string, value []GenericZ[T]) error {
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
	if err := p.db.ZAdd(ctx, key, zSlice...).Err(); err != nil {
		return fmt.Errorf("failed to push to priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) PopMin(ctx context.Context, key string) error {
	if err := p.db.ZPopMin(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to pop min from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) PopMax(ctx context.Context, key string) error {
	if err := p.db.ZPopMax(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to pop max from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) DeleteRange(ctx context.Context, key string, offset, limit int64) error {
	if err := p.db.ZRemRangeByRank(ctx, key, offset, limit).Err(); err != nil {
		return fmt.Errorf("failed to delete range from priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) Peek(ctx context.Context, key string) (T, error) {
	var val T
	valStrs, err := p.db.ZRange(ctx, key, 0, 0).Result()
	if err != nil {
		return val, fmt.Errorf("failed to peek priority queue key %s: %w", key, err)
	}
	if len(valStrs) == 0 {
		return val, nil
	}
	return p.unmarshalOne(ctx, valStrs[0])
}

func (p *PriorityQueue[T]) IsEmpty(ctx context.Context, key string) (bool, error) {
	sz, err := p.Size(ctx, key)
	return sz == 0, err
}

func (p *PriorityQueue[T]) Size(ctx context.Context, key string) (int64, error) {
	result, err := p.db.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get size of priority queue key %s: %w", key, err)
	}
	return result, nil
}

func (p *PriorityQueue[T]) Clear(ctx context.Context, key string) error {
	if err := p.db.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear priority queue key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) ChangePriority(ctx context.Context, key string, value T, priority int64) error {
	if err := p.db.ZAdd(ctx, key, redis.Z{Score: float64(priority), Member: value}).Err(); err != nil {
		return fmt.Errorf("failed to change priority for key %s: %w", key, err)
	}
	return nil
}

func (p *PriorityQueue[T]) ToArray(ctx context.Context, key string) ([]T, error) {
	valStrs, err := p.db.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to convert priority queue to array for key %s: %w", key, err)
	}
	return p.unmarshalSlice(ctx, valStrs)
}

func (p *PriorityQueue[T]) unmarshalOne(ctx context.Context, s string) (T, error) {
	var val T
	cmd := redis.NewStringCmd(ctx)
	cmd.SetVal(s)
	err := cmd.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan priority queue value: %w", err)
	}
	return val, nil
}

func (p *PriorityQueue[T]) unmarshalSlice(ctx context.Context, valStrs []string) ([]T, error) {
	res := make([]T, len(valStrs))
	for i, s := range valStrs {
		v, err := p.unmarshalOne(ctx, s)
		if err != nil {
			return nil, err
		}
		res[i] = v
	}
	return res, nil
}
