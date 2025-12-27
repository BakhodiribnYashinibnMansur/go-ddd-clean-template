package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type String struct {
	db *redis.Client
}

func NewString(db *redis.Client) *String {
	return &String{
		db: db,
	}
}
func (s *String) Get(key string) (string, error) {
	value, err := s.db.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}
func (s *String) Pop(key string) (string, error) {
	value, err := s.db.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	err = s.db.Del(context.Background(), key).Err()
	if err != nil {
		return "", err
	}
	return value, nil
}
func (s *String) Set(key string, value string, expiration time.Duration) error {
	err := s.db.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *String) Delete(key string) error {
	err := s.db.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *String) Scan(pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = s.db.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (s *String) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := s.db.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}
