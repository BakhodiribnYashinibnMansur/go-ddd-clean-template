package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type PriorityQueue struct {
	db *redis.Client
}

func NewPriorityQueue(db *redis.Client) *PriorityQueue {
	return &PriorityQueue{
		db: db,
	}
}

func (p *PriorityQueue) Get(key string, offset, limit int64) ([]string, error) {
	value, err := p.db.ZRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (p *PriorityQueue) GetFull(key string) (int64, error) {
	value, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(value)), nil
}

func (p *PriorityQueue) Delete(key string) error {
	err := p.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) Push(key string, value []redis.Z) error {
	err := p.db.ZAdd(context.Background(), key, value...).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) PopMin(key string) error {
	err := p.db.ZPopMin(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) PopMax(key string) error {
	err := p.db.ZPopMax(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) DeleteRange(key string, offset, limit int64) error {
	err := p.db.ZRemRangeByRank(context.Background(), key, offset, limit).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) Peek(key string) (string, error) {
	value, err := p.db.ZRange(context.Background(), key, 0, 0).Result()
	if err != nil {
		return "", err
	}
	return value[0], nil
}

func (p *PriorityQueue) IsEmpty(key string) (bool, error) {
	value, err := p.db.ZRange(context.Background(), key, 0, 0).Result()
	if err != nil {
		return false, err
	}
	return len(value) == 0, nil
}

func (p *PriorityQueue) Size(key string) (int64, error) {
	value, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(value)), nil
}

func (p *PriorityQueue) Clear(key string) error {
	err := p.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) ChangePriority(key string, value string, priority int64) error {
	values, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	err = p.db.ZAdd(context.Background(), key, redis.Z{Score: float64(priority), Member: value}).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PriorityQueue) ToArray(key string) ([]string, error) {
	value, err := p.db.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}
