package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	db *redis.Client
}

func NewQueue(db *redis.Client) *Queue {
	return &Queue{
		db: db,
	}
}

func (q *Queue) Len(key string) (int64, error) {
	value, err := q.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (q *Queue) Get(key string, offset, limit int64) (items []string, err error) {
	items, err = q.db.LRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queue) GetFull(key string) (int64, error) {
	items, err := q.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (q *Queue) Delete(key string) error {
	err := q.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) Pop(key string) error {
	err := q.db.RPop(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) PushFront(key string, value []any) error {
	for _, v := range value {
		err := q.db.LPush(context.Background(), key, v).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) PushBack(key string, value []any) error {
	for _, v := range value {
		err := q.db.RPush(context.Background(), key, v).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) PopFront(key string) (string, error) {
	value, err := q.db.LPop(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (q *Queue) PopBack(key string) (string, error) {
	value, err := q.db.RPop(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (q *Queue) DeleteRange(key string, offset, limit int64) error {
	err := q.db.LTrim(context.Background(), key, offset, limit).Err()
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) Peek(key string) (string, error) {
	value, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (q *Queue) IsEmpty(key string) (bool, error) {
	value, err := q.db.LLen(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return value == 0, nil
}

func (q *Queue) Contains(key string, valueIn string) (bool, error) {
	valueOut, err := q.db.LIndex(context.Background(), key, 0).Result()
	if err != nil {
		return false, err
	}
	return valueIn == valueOut, nil
}

func (q *Queue) ToArray(key string) ([]string, error) {
	value, err := q.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}
