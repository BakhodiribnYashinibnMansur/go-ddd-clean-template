package integration

import (
	"context"
	"fmt"
	"time"

	"gct/internal/domain"
)

// ValidateAPIKey business logic using in-memory cache.
func (uc *UseCase) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	// First hash the provided key to look it up in the cache.
	tempKey := &domain.APIKey{}
	hashedKey := tempKey.HashKey(key)

	uc.mu.RLock()
	apiKey, exists := uc.apiKeys[hashedKey]
	uc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid api key")
	}

	if !apiKey.IsActive {
		return nil, fmt.Errorf("api key is inactive")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("api key expired")
	}

	// Update last used timestamp asynchronously
	go func() {
		_ = uc.repo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
	}()

	return apiKey, nil
}
