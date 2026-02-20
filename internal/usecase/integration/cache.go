package integration

import (
	"context"

	"gct/consts"
	"gct/internal/domain"

	"github.com/google/uuid"
)

// InitCache loads all active integrations and API keys from database into memory.
func (uc *UseCase) InitCache(ctx context.Context) error {
	uc.logger.Infoc(ctx, "🔄 Initializing integration and API key cache...")

	// 1. Load integrations
	integrations, _, err := uc.repo.ListIntegrations(ctx, domain.IntegrationFilter{
		IsActive: Ptr(true),
	})
	if err != nil {
		return err
	}

	// 2. Load API keys and map them
	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Clear existing cache
	uc.integrations = make(map[uuid.UUID]*domain.Integration)
	uc.apiKeys = make(map[string]*domain.APIKey)

	for i := range integrations {
		integration := integrations[i]
		uc.integrations[integration.ID] = &integration

		// List keys for this integration
		keys, err := uc.repo.ListAPIKeysByIntegration(ctx, integration.ID)
		if err != nil {
			uc.logger.Errorc(ctx, "failed to load api keys for integration", "error", err, "integration_id", integration.ID)
			continue
		}

		for k := range keys {
			if keys[k].IsActive {
				apiKey := keys[k]
				uc.apiKeys[apiKey.Key] = &apiKey
			}
		}
	}

	uc.logger.Infoc(ctx, "✅ Integration cache initialized",
		"integrations_count", len(uc.integrations),
		"api_keys_count", len(uc.apiKeys),
	)
	return nil
}

// InvalidateCache refreshes the cache when a database change is detected.
func (uc *UseCase) InvalidateCache(ctx context.Context, table string) error {
	uc.logger.Infoc(ctx, "♻️ Invalidating integration cache due to change in table", "table", table)

	// We could be smart and only update changed records, but for now re-initializing is simpler and safer
	// given the likely small number of integrations and API keys.
	if table == consts.TableIntegrations || table == consts.TableAPIKeys {
		if err := uc.InitCache(ctx); err != nil {
			uc.logger.Errorc(ctx, "failed to re-initialize integration cache", "error", err)
			return err
		}
	}
	return nil
}

// Ptr helper (if not already everywhere, but used here)
func Ptr[T any](v T) *T {
	return &v
}
