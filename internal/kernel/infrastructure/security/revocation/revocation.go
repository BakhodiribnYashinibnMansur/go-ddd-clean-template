package revocation

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "jwt:revoked:"

// RedisClient is the minimal interface needed (satisfied by *redis.Client).
type RedisClient interface {
	SetEx(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Pipeline() redis.Pipeliner
}

// Store manages a Redis-backed denylist of revoked session IDs.
// All methods are safe for concurrent use.
type Store struct {
	client RedisClient
}

// New creates a new revocation Store.
func New(client RedisClient) *Store {
	return &Store{client: client}
}

// Revoke marks a session as revoked. ttl should be the remaining lifetime
// of the access token (so the entry auto-expires when the token would have
// expired naturally).
func (s *Store) Revoke(ctx context.Context, sessionID string, ttl time.Duration) error {
	return s.client.SetEx(ctx, keyPrefix+sessionID, "1", ttl).Err()
}

// IsRevoked returns true if the session has been revoked.
// On Redis errors it returns false (fail-open).
func (s *Store) IsRevoked(ctx context.Context, sessionID string) bool {
	result, err := s.client.Exists(ctx, keyPrefix+sessionID).Result()
	if err != nil {
		return false
	}
	return result > 0
}

// RevokeMany marks multiple sessions as revoked in a pipeline.
func (s *Store) RevokeMany(ctx context.Context, sessionIDs []string, ttl time.Duration) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, id := range sessionIDs {
		pipe.SetEx(ctx, keyPrefix+id, "1", ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}
