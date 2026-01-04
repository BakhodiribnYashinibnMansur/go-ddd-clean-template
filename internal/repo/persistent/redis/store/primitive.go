package store

import (
	"context"
	"errors"
	"fmt"
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
		return val, fmt.Errorf("failed to get primitive value from key %s: %w", key, err)
	}
	return val, nil
}

func (p *Primitive[T]) Set(key string, value T, expiration time.Duration) error {
	if err := p.db.Set(context.Background(), key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set primitive value to key %s: %w", key, err)
	}
	return nil
}

func (p *Primitive[T]) Delete(key string) error {
	if err := p.db.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete primitive key %s: %w", key, err)
	}
	return nil
}

func (p *Primitive[T]) Pop(key string) (T, error) {
	var val T
	ctx := context.Background()
	pipe := p.db.TxPipeline()
	get := pipe.Get(ctx, key)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return val, fmt.Errorf("failed to execute pipeline for primitive key %s: %w", key, err)
	}

	err = get.Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to scan primitive value: %w", err)
	}
	return val, nil
}

func (p *Primitive[T]) Exists(key string) (bool, error) {
	val, err := p.db.Exists(context.Background(), key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of primitive key %s: %w", key, err)
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
			return nil, fmt.Errorf("failed to scan primitive keys with pattern %s: %w", pattern, err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
