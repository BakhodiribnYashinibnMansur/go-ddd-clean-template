package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrIPRateLimited is returned when an IP exceeds the allowed number of
	// failed authentication attempts within the sliding window.
	ErrIPRateLimited = errors.New("too many requests from this IP")

	// ErrAccountLocked is returned when a user account is temporarily locked
	// due to repeated failed authentication attempts.
	ErrAccountLocked = errors.New("account temporarily locked")
)

// RedisClient is the minimal Redis interface required by AuthLimiter.
type RedisClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
}

// Config holds tunables for the authentication rate limiter.
type Config struct {
	IPLimit     int           // max failed attempts per IP per window
	IPWindow    time.Duration // sliding window duration for IP tracking
	UserLimit   int           // max failed attempts per user before lockout
	UserWindow  time.Duration // sliding window duration for user tracking
	LockoutBase time.Duration // initial lockout duration
}

// DefaultConfig returns production-ready defaults.
func DefaultConfig() Config {
	return Config{
		IPLimit:     10,
		IPWindow:    60 * time.Second,
		UserLimit:   5,
		UserWindow:  15 * time.Minute,
		LockoutBase: 30 * time.Minute,
	}
}

// AuthLimiter provides rate limiting for authentication endpoints.
// It tracks failed attempts per IP and per login identifier separately.
type AuthLimiter struct {
	client      RedisClient
	ipLimit     int
	ipWindow    time.Duration
	userLimit   int
	userWindow  time.Duration
	lockoutBase time.Duration
}

// New creates an AuthLimiter with the given Redis client and configuration.
func New(client RedisClient, cfg Config) *AuthLimiter {
	return &AuthLimiter{
		client:      client,
		ipLimit:     cfg.IPLimit,
		ipWindow:    cfg.IPWindow,
		userLimit:   cfg.UserLimit,
		userWindow:  cfg.UserWindow,
		lockoutBase: cfg.LockoutBase,
	}
}

// --- key helpers ---

func ipFailKey(ip string) string       { return fmt.Sprintf("auth:ip_fail:%s", ip) }
func userFailKey(login string) string   { return fmt.Sprintf("auth:user_fail:%s", login) }
func userLockKey(login string) string   { return fmt.Sprintf("auth:user_lock:%s", login) }
func userLockMultKey(login string) string { return fmt.Sprintf("auth:user_lock_mult:%s", login) }

// CheckIP returns ErrIPRateLimited if the IP has exceeded the allowed number
// of failed attempts within the sliding window.
func (l *AuthLimiter) CheckIP(ctx context.Context, ip string) error {
	val, err := l.client.Get(ctx, ipFailKey(ip)).Int64()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		// On Redis errors, allow the request (fail open).
		return nil
	}
	if val >= int64(l.ipLimit) {
		return ErrIPRateLimited
	}
	return nil
}

// CheckUser returns ErrAccountLocked if the user account is currently locked.
func (l *AuthLimiter) CheckUser(ctx context.Context, login string) error {
	_, err := l.client.Get(ctx, userLockKey(login)).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return nil
	}
	return ErrAccountLocked
}

// RecordFailedIP increments the failed counter for an IP address.
// The counter uses a simple INCR + EXPIRE sliding window.
func (l *AuthLimiter) RecordFailedIP(ctx context.Context, ip string) error {
	key := ipFailKey(ip)

	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("incr %s: %w", key, err)
	}

	// Set TTL on first increment.
	if count == 1 {
		if err := l.client.Expire(ctx, key, l.ipWindow).Err(); err != nil {
			return fmt.Errorf("expire %s: %w", key, err)
		}
	}

	return nil
}

// RecordFailedUser increments the failed counter for a login identifier.
// When the threshold is exceeded, a lockout is applied with exponential
// backoff (30m -> 60m -> 120m ...).
func (l *AuthLimiter) RecordFailedUser(ctx context.Context, login string) error {
	failKey := userFailKey(login)

	count, err := l.client.Incr(ctx, failKey).Result()
	if err != nil {
		return fmt.Errorf("incr %s: %w", failKey, err)
	}

	if count == 1 {
		if err := l.client.Expire(ctx, failKey, l.userWindow).Err(); err != nil {
			return fmt.Errorf("expire %s: %w", failKey, err)
		}
	}

	if count >= int64(l.userLimit) {
		if err := l.applyLockout(ctx, login); err != nil {
			return err
		}
	}

	return nil
}

// applyLockout sets the lockout key with exponential backoff and advances the
// multiplier.
func (l *AuthLimiter) applyLockout(ctx context.Context, login string) error {
	multKey := userLockMultKey(login)

	mult := int64(1)
	if v, err := l.client.Get(ctx, multKey).Result(); err == nil {
		if parsed, pErr := strconv.ParseInt(v, 10, 64); pErr == nil && parsed > 0 {
			mult = parsed
		}
	}

	lockDuration := l.lockoutBase * time.Duration(mult)

	if err := l.client.Set(ctx, userLockKey(login), "1", lockDuration).Err(); err != nil {
		return fmt.Errorf("set lockout: %w", err)
	}

	// Double the multiplier for next lockout, retain for 24h.
	nextMult := mult * 2
	if err := l.client.Set(ctx, multKey, strconv.FormatInt(nextMult, 10), 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("set lock multiplier: %w", err)
	}

	// Clear the fail counter so it can start fresh after lockout expires.
	if err := l.client.Del(ctx, userFailKey(login)).Err(); err != nil {
		return fmt.Errorf("del fail counter: %w", err)
	}

	return nil
}

// ResetUser clears the failed counter and lockout state for a login
// identifier. Call this on successful authentication.
func (l *AuthLimiter) ResetUser(ctx context.Context, login string) error {
	keys := []string{
		userFailKey(login),
		userLockKey(login),
		userLockMultKey(login),
	}
	if err := l.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("reset user rate limit state: %w", err)
	}
	return nil
}
