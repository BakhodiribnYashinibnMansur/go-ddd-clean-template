package csrf

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCSRFTokenNotFound is returned when CSRF token is not found in store.
var ErrCSRFTokenNotFound = errors.New("CSRF token not found")

// RedisStore implements stateful CSRF token storage using Redis
// Suitable for production multi-instance deployments
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore creates a new Redis-based CSRF store
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: "csrf:",
	}
}

// Set stores a CSRF token hash with expiration
func (s *RedisStore) Set(ctx context.Context, sessionID, tokenHash string, expiration time.Duration) error {
	key := s.key(sessionID)

	err := s.client.Set(ctx, key, tokenHash, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis csrf set failed: %w", err)
	}

	return nil
}

// Get retrieves a CSRF token hash and its expiration
func (s *RedisStore) Get(ctx context.Context, sessionID string) (string, time.Time, error) {
	key := s.key(sessionID)

	// Get token hash
	tokenHash, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", time.Time{}, fmt.Errorf("%w for session: %s", ErrCSRFTokenNotFound, sessionID)
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("redis csrf get failed: %w", err)
	}

	// Get TTL for expiration time
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("redis csrf ttl failed: %w", err)
	}

	expiresAt := time.Now().Add(ttl)

	return tokenHash, expiresAt, nil
}

// Delete removes a CSRF token
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	key := s.key(sessionID)

	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis csrf delete failed: %w", err)
	}

	return nil
}

// Rotate replaces an old token with a new one atomically
// Uses Redis transaction for atomicity
func (s *RedisStore) Rotate(ctx context.Context, sessionID, newTokenHash string, expiration time.Duration) error {
	key := s.key(sessionID)

	// Use transaction for atomic rotation
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, key)
	pipe.Set(ctx, key, newTokenHash, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis csrf rotate failed: %w", err)
	}

	return nil
}

// key generates Redis key with prefix
func (s *RedisStore) key(sessionID string) string {
	return s.prefix + sessionID
}

// Count returns the number of stored CSRF tokens (for monitoring)
func (s *RedisStore) Count(ctx context.Context) (int64, error) {
	keys, err := s.client.Keys(ctx, s.prefix+"*").Result()
	if err != nil {
		return 0, fmt.Errorf("redis csrf count failed: %w", err)
	}

	return int64(len(keys)), nil
}

// Cleanup removes expired tokens (Redis handles this automatically via TTL)
// This method is provided for compatibility but is a no-op for Redis
func (s *RedisStore) Cleanup(ctx context.Context) error {
	// Redis automatically removes expired keys
	// No manual cleanup needed
	return nil
}
