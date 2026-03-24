package clickhouse

import (
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Option defines a function type for configuring ClickHouse connection.
type Option func(*clickhouse.Options)

// WithDialTimeout sets the dial timeout.
func WithDialTimeout(d time.Duration) Option {
	return func(opts *clickhouse.Options) {
		opts.DialTimeout = d
	}
}

// WithCompression enables/disables compression.
func WithCompression(c *clickhouse.Compression) Option {
	return func(opts *clickhouse.Options) {
		opts.Compression = c
	}
}

// WithDebug enables debug mode.
func WithDebug(debug bool) Option {
	return func(opts *clickhouse.Options) {
		opts.Debug = debug
	}
}
