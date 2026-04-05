// Package redis implements Redis connection.
package redis

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

const (
	// ENV PROD
	poolSizeProd     = 50
	minIdleConnsProd = 10

	// ENV DEV
	poolSizeDev     = 8
	minIdleConnsDev = 3

	defaultConnMaxLifetime = 10 * time.Hour
	defaultConnMaxIdleTime = 30 * time.Minute
	defaultDialTimeout     = 10 * time.Second
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 5 * time.Second
	defaultPingTimeout     = 5 * time.Second
)

// Redis struct wraps redis.Client.
type Redis struct {
	Client *redis.Client
}

// New creates a new Redis client with optimized settings.
func New(ctx context.Context, env string, cfg config.Redis, l logger.Log, opts ...Option) (*Redis, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	options := &redis.Options{
		Addr:         addr,
		Username:     cfg.User,
		Password:     cfg.Password,
		DB:           0, // default DB
		DialTimeout:  defaultDialTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	// Apply pool configuration
	applyPoolConfig(env, options)

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	client := redis.NewClient(options)

	// Enable tracing
	if err := redisotel.InstrumentTracing(client); err != nil {
		l.Errorc(ctx, "failed to instrument redis with otel", "error", err)
		// We don't return error here to avoid breaking app if tracing fails
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, client, l); err != nil {
		client.Close()
		return nil, err
	}

	r := &Redis{
		Client: client,
	}

	clientOpts := client.Options()
	l.Infoc(ctx, "✅ Redis connected successfully",
		"address", clientOpts.Addr,
		"username", clientOpts.Username,
		"db", clientOpts.DB,
		"pool_size", clientOpts.PoolSize,
		"min_idle_conns", clientOpts.MinIdleConns,
	)

	return r, nil
}

// applyPoolConfig applies pool configuration with defaults.
func applyPoolConfig(env string, options *redis.Options) {
	if env == "production" || env == "PROD" {
		options.PoolSize = poolSizeProd
		options.MinIdleConns = minIdleConnsProd
	} else {
		options.PoolSize = poolSizeDev
		options.MinIdleConns = minIdleConnsDev
	}

	options.ConnMaxLifetime = defaultConnMaxLifetime
	options.ConnMaxIdleTime = defaultConnMaxIdleTime
}

// verifyConnection pings the Redis server to ensure connectivity.
func verifyConnection(ctx context.Context, client *redis.Client, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		opts := client.Options()
		l.Errorc(ctx, "❌ Redis connection failed",
			"error", err,
			"address", opts.Addr,
			"username", opts.Username,
			"db", opts.DB,
			"dial_timeout", opts.DialTimeout,
			"read_timeout", opts.ReadTimeout,
			"write_timeout", opts.WriteTimeout,
		)
		return fmt.Errorf("verify redis connection: %w", err)
	}

	return nil
}

// Close gracefully closes the Redis connection.
func (r *Redis) Close() error {
	if r != nil && r.Client != nil {
		if err := r.Client.Close(); err != nil {
			return fmt.Errorf("redis.Close: %w", err)
		}
	}
	return nil
}

// Stats returns current pool statistics.
func (r *Redis) Stats() *redis.PoolStats {
	if r == nil || r.Client == nil {
		return nil
	}
	return r.Client.PoolStats()
}
