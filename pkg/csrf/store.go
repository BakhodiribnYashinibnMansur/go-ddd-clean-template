package csrf

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Store defines the interface for CSRF token storage
type Store interface {
	// Set stores a CSRF token hash with expiration
	Set(ctx context.Context, sessionID, tokenHash string, expiration time.Duration) error

	// Get retrieves a CSRF token hash and its expiration
	Get(ctx context.Context, sessionID string) (tokenHash string, expiresAt time.Time, err error)

	// Delete removes a CSRF token
	Delete(ctx context.Context, sessionID string) error

	// Rotate replaces an old token with a new one atomically
	Rotate(ctx context.Context, sessionID, newTokenHash string, expiration time.Duration) error
}

// MemoryStore implements in-memory CSRF token storage
// Suitable for development or single-instance deployments
type MemoryStore struct {
	mu     sync.RWMutex
	tokens map[string]*storedToken
}

type storedToken struct {
	hash      string
	expiresAt time.Time
}

// NewMemoryStore creates a new in-memory CSRF store
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		tokens: make(map[string]*storedToken),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Set stores a CSRF token hash
func (s *MemoryStore) Set(ctx context.Context, sessionID, tokenHash string, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[sessionID] = &storedToken{
		hash:      tokenHash,
		expiresAt: time.Now().Add(expiration),
	}

	return nil
}

// Get retrieves a CSRF token hash
func (s *MemoryStore) Get(ctx context.Context, sessionID string) (string, time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[sessionID]
	if !exists {
		return "", time.Time{}, fmt.Errorf("%w for session: %s", ErrCSRFTokenNotFound, sessionID)
	}

	// Check if expired
	if time.Now().After(token.expiresAt) {
		return "", time.Time{}, ErrExpiredToken
	}

	return token.hash, token.expiresAt, nil
}

// Delete removes a CSRF token
func (s *MemoryStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, sessionID)
	return nil
}

// Rotate replaces an old token with a new one
func (s *MemoryStore) Rotate(ctx context.Context, sessionID, newTokenHash string, expiration time.Duration) error {
	return s.Set(ctx, sessionID, newTokenHash, expiration)
}

// cleanup periodically removes expired tokens
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for sessionID, token := range s.tokens {
			if now.After(token.expiresAt) {
				delete(s.tokens, sessionID)
			}
		}
		s.mu.Unlock()
	}
}

// Count returns the number of stored tokens (for testing)
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tokens)
}
