package apikeythrottle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrThrottled is returned when an IP has exceeded the failure limit.
var ErrThrottled = errors.New("too many invalid API key attempts")

const (
	keyPrefix  = "jwt:apikey_fail:"
	blockPrefix = "jwt:apikey_block:"
	defaultLimit  = 20
	defaultWindow = 10 * time.Minute
	defaultBlock  = 1 * time.Hour
)

// Config controls throttle behaviour.
type Config struct {
	Limit  int           // max bad attempts per IP per window (default 20)
	Window time.Duration // sliding window (default 10m)
	Block  time.Duration // block duration after limit exceeded (default 1h)
}

// DefaultConfig returns the default throttle configuration.
func DefaultConfig() Config {
	return Config{
		Limit:  defaultLimit,
		Window: defaultWindow,
		Block:  defaultBlock,
	}
}

// Throttle tracks failed API key attempts per IP via Redis.
type Throttle struct {
	client redis.Cmdable
	cfg    Config
}

// New creates a Throttle with the given Redis client and config.
func New(client redis.Cmdable, cfg Config) *Throttle {
	if cfg.Limit == 0 {
		cfg.Limit = defaultLimit
	}
	if cfg.Window == 0 {
		cfg.Window = defaultWindow
	}
	if cfg.Block == 0 {
		cfg.Block = defaultBlock
	}

	return &Throttle{client: client, cfg: cfg}
}

// Check returns ErrThrottled if the IP is currently blocked.
func (t *Throttle) Check(ctx context.Context, ip string) error {
	blocked, err := t.client.Exists(ctx, blockPrefix+ip).Result()
	if err != nil {
		return fmt.Errorf("apikeythrottle check: %w", err)
	}

	if blocked > 0 {
		return ErrThrottled
	}

	return nil
}

// RecordFail increments the failure counter. If it exceeds the limit,
// sets a block key.
func (t *Throttle) RecordFail(ctx context.Context, ip string) error {
	failKey := keyPrefix + ip

	count, err := t.client.Incr(ctx, failKey).Result()
	if err != nil {
		return fmt.Errorf("apikeythrottle incr: %w", err)
	}

	if count == 1 {
		t.client.Expire(ctx, failKey, t.cfg.Window)
	}

	if int(count) >= t.cfg.Limit {
		t.client.Set(ctx, blockPrefix+ip, "1", t.cfg.Block)
	}

	return nil
}
