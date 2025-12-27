package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type List struct {
	db *redis.Client
}

func NewList(db *redis.Client) *List {
	return &List{
		db: db,
	}
}

func (l *List) Get(key string) ([]string, error) {
	value, err := l.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}
func (l *List) GetFull(key string) (int64, error) {
	listLen, err := l.db.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return 0, err
	}
	return int64(len(listLen)), nil
}

func (l *List) Pop(key string, offset, limit int64) ([]string, error) {
	set, err := l.db.LRange(context.Background(), key, offset, limit).Result()
	if err != nil {
		return nil, err
	}
	err = l.db.Del(context.Background(), key).Err()
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (l *List) Set(key string, setKey []any, expiration time.Duration) error {
	err := l.db.LPush(context.Background(), key, setKey...).Err()
	if err != nil {
		return err
	}

	err = l.db.Expire(context.Background(), key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (l *List) Delete(key string) error {
	err := l.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
func (l *List) Len(key string) (int64, error) {
	len, err := l.db.LLen(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	return len, nil
}
