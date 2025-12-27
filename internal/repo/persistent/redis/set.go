package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Set struct {
	db *redis.Client
}

func NewSet(db *redis.Client) *Set {
	return &Set{
		db: db,
	}
}

func (s *Set) Get(key string) ([]string, error) {
	value, err := s.db.SMembers(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Set) Pop(key string) ([]string, error) {
	set, err := s.db.SInter(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	err = s.db.Del(context.Background(), key).Err()
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (s *Set) Set(key string, setKey []any, expiration time.Duration) error {
	err := s.db.SAdd(context.Background(), key, setKey).Err()
	if err != nil {
		return err
	}
	err = s.db.Expire(context.Background(), key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *Set) Delete(key string) error {
	err := s.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
