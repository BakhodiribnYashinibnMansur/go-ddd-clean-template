package store

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNoUnmarshaller = errors.New("no unmarshaller defined")

// PrimitiveI defines basic CRUD operations for any primitive type T
type PrimitiveI[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T, expiration time.Duration) error
	Delete(key string) error
	Pop(key string) (T, error)
	Exists(key string) (bool, error)
	Scan(pattern string) ([]string, error)
}

// Primitives contains all primitive types
type Primitives struct {
	String PrimitiveI[string]
	Int    PrimitiveI[int64]
	Byte   PrimitiveI[[]byte]
	Bool   PrimitiveI[bool]
	Float  PrimitiveI[float64]
}

// Primitive provides basic CRUD operations for any type T
type Primitive[T any] struct {
	db *redis.Client
}

// NewPrimitive creates a new Primitive store
func NewPrimitive[T any](db *redis.Client) *Primitive[T] {
	return &Primitive[T]{
		db: db,
	}
}

func (p *Primitive[T]) Get(key string) (T, error) {
	var val T
	err := p.db.Get(context.Background(), key).Scan(&val)
	if err != nil {
		return val, err
	}
	return val, nil
}

func (p *Primitive[T]) Set(key string, value T, expiration time.Duration) error {
	return p.db.Set(context.Background(), key, value, expiration).Err()
}

func (p *Primitive[T]) Delete(key string) error {
	return p.db.Del(context.Background(), key).Err()
}

func (p *Primitive[T]) Pop(key string) (T, error) {
	var val T
	ctx := context.Background()
	pipe := p.db.TxPipeline()
	get := pipe.Get(ctx, key)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return val, err
	}

	err = get.Scan(&val)
	return val, err
}

func (p *Primitive[T]) Exists(key string) (bool, error) {
	val, err := p.db.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (p *Primitive[T]) Scan(pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = p.db.Scan(context.Background(), cursor, pattern, 100).Result()
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
