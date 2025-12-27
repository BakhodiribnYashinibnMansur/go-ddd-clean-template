package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Option defines a function type for configuring Redis client.
type Option func(*redis.Options)

// WithPoolSize sets the maximum number of socket connections.
func WithPoolSize(n int) Option {
	return func(opts *redis.Options) {
		opts.PoolSize = n
	}
}

// WithMinIdleConns sets the minimum number of idle connections.
func WithMinIdleConns(n int) Option {
	return func(opts *redis.Options) {
		opts.MinIdleConns = n
	}
}

// WithDB sets the database to use.
func WithDB(db int) Option {
	return func(opts *redis.Options) {
		opts.DB = db
	}
}

// WithDialTimeout sets the timeout for establishing new connections.
func WithDialTimeout(d time.Duration) Option {
	return func(opts *redis.Options) {
		opts.DialTimeout = d
	}
}

// WithReadTimeout sets the timeout for socket reads.
func WithReadTimeout(d time.Duration) Option {
	return func(opts *redis.Options) {
		opts.ReadTimeout = d
	}
}

// WithWriteTimeout sets the timeout for socket writes.
func WithWriteTimeout(d time.Duration) Option {
	return func(opts *redis.Options) {
		opts.WriteTimeout = d
	}
}
