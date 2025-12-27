// Package redis implements Redis connection.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
func New(ctx context.Context, env string, cfg config.Redis, l logger.Interface, opts ...Option) (*Redis, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	options := &redis.Options{
		Addr:         addr,
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

	// Verify connection with ping
	if err := verifyConnection(ctx, client, l); err != nil {
		client.Close()
		return nil, err
	}

	r := &Redis{
		Client: client,
	}

	l.Infow("Redis connected successfully")

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
func verifyConnection(ctx context.Context, client *redis.Client, l logger.Interface) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		l.Errorw("failed to ping Redis server", zap.Error(err))
		return fmt.Errorf("verify redis connection: %w", err)
	}

	return nil
}

// Close gracefully closes the Redis connection.
func (r *Redis) Close() error {
	if r != nil && r.Client != nil {
		return r.Client.Close()
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
