package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Option defines a function type for configuring MongoDB client.
type Option func(*options.ClientOptions)

// WithMaxPoolSize sets the maximum number of connections in the pool.
func WithMaxPoolSize(n uint64) Option {
	return func(opts *options.ClientOptions) {
		opts.SetMaxPoolSize(n)
	}
}

// WithMinPoolSize sets the minimum number of connections in the pool.
func WithMinPoolSize(n uint64) Option {
	return func(opts *options.ClientOptions) {
		opts.SetMinPoolSize(n)
	}
}

// WithMaxConnIdleTime sets the maximum idle time for connections.
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(opts *options.ClientOptions) {
		opts.SetMaxConnIdleTime(d)
	}
}

// WithTimeout sets the connection timeout.
func WithTimeout(d time.Duration) Option {
	return func(opts *options.ClientOptions) {
		opts.SetConnectTimeout(d)
	}
}
