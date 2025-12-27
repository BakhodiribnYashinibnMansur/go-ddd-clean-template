package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	StringI
	ArrayI
	HashTableI
	SetI
	IntI
	QueueI
	ListI
	ByteI
	BoolI
}

type StringI interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) (string, error)
	Scan(pattern string) ([]string, error)
	Exists(key string) (bool, error)
}

type ArrayI interface {
	Get(key string) ([]string, error)
	Set(key string, value []string, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) ([]string, error)
}

type HashTableI interface {
	Get(key string, delete bool) (map[string]string, error)
	Pop(key string) (map[string]string, error)
	Set(key string, hashKey map[string]any, expirationTime time.Duration) error
	Delete(key string) error
}

type SetI interface {
	Get(key string) ([]string, error)
	Set(key string, value []any, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) ([]string, error)
}

type IntI interface {
	Get(key string) (int64, error)
	Set(key string, value int64, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) (int64, error)
}

type QueueI interface {
	Len(key string) (int64, error)
	Get(key string, offset, limit int64) ([]string, error)
	GetFull(key string) (int64, error)
	Delete(key string) error
	Pop(key string) error
	PushFront(key string, value []any) error
	PushBack(key string, value []any) error
	PopFront(key string) (string, error)
	PopBack(key string) (string, error)
	DeleteRange(key string, offset, limit int64) error
	Peek(key string) (string, error)
	IsEmpty(key string) (bool, error)
	Contains(key string, value string) (bool, error)
	ToArray(key string) ([]string, error)
}

type ListI interface {
	Get(key string) ([]string, error)
	Set(key string, value []any, expiration time.Duration) error
	Delete(key string) error
	Pop(key string, limit int64, offset int64) ([]string, error)
	GetFull(key string) (int64, error)
	Len(key string) (int64, error)
}

type ByteI interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, expiration time.Duration) error
	Delete(key string) error
}

type BoolI interface {
	Get(key string) (bool, error)
	Set(key string, value bool, expiration time.Duration) error
	Delete(key string) error
}

func NewRedisRepo(redis *redis.Client) *RedisRepo {
	return &RedisRepo{
		StringI:    NewString(redis),
		ArrayI:     nil, // ArrayI implementation not found
		SetI:       NewSet(redis),
		IntI:       NewInt(redis),
		QueueI:     NewQueue(redis),
		ListI:      NewList(redis),
		HashTableI: NewHashTable(redis),
		ByteI:      NewByte(redis),
		BoolI:      NewBool(redis),
	}
}
