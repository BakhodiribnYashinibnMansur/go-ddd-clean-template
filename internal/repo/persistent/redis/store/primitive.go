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
	Get(ctx context.Context, key string) (T, error)
	Set(ctx context.Context, key string, value T, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Pop(ctx context.Context, key string) (T, error)
	Exists(ctx context.Context, key string) (bool, error)
	Scan(ctx context.Context, pattern string) ([]string, error)
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

func (p *Primitive[T]) Get(ctx context.Context, key string) (T, error) {
	var val T
	err := p.db.Get(ctx, key).Scan(&val)
	if err != nil {
		return val, fmt.Errorf("failed to get primitive value from key %s: %w", key, err)
	}
	return val, nil
}

func (p *Primitive[T]) Set(ctx context.Context, key string, value T, expiration time.Duration) error {
	if err := p.db.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set primitive value to key %s: %w", key, err)
	}
	return nil
}

func (p *Primitive[T]) Delete(ctx context.Context, key string) error {
	if err := p.db.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete primitive key %s: %w", key, err)
	}
	return nil
}

func (p *Primitive[T]) Pop(ctx context.Context, key string) (T, error) {
	var val T
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

func (p *Primitive[T]) Exists(ctx context.Context, key string) (bool, error) {
	val, err := p.db.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of primitive key %s: %w", key, err)
	}
	return val > 0, nil
}

func (p *Primitive[T]) Scan(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = p.db.Scan(ctx, cursor, pattern, 100).Result()
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
