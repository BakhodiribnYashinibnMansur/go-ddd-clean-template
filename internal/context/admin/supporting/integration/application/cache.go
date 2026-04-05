package application

import (
	"context"
	"fmt"
	"sync"

	"gct/internal/context/admin/supporting/integration/domain"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/logger"
)

// CachedIntegration holds an in-memory snapshot of an active integration.
type CachedIntegration struct {
	ID         domain.IntegrationID
	Name       string
	Type       string
	APIKey     string
	WebhookURL string
	Enabled    bool
	Config     map[string]string
}

// CacheService manages an in-memory cache of active integrations and API keys for fast lookup.
type CacheService struct {
	readRepo domain.IntegrationReadRepository
	logger   logger.Log

	mu           sync.RWMutex
	integrations map[domain.IntegrationID]*CachedIntegration
	apiKeys      map[string]*CachedIntegration // keyed by API key string
}

// NewCacheService creates a new integration cache service.
func NewCacheService(readRepo domain.IntegrationReadRepository, l logger.Log) *CacheService {
	return &CacheService{
		readRepo:     readRepo,
		logger:       l,
		integrations: make(map[domain.IntegrationID]*CachedIntegration),
		apiKeys:      make(map[string]*CachedIntegration),
	}
}

// InitCache loads all enabled integrations from the database into memory.
func (s *CacheService) InitCache(ctx context.Context) error {
	s.logger.Infoc(ctx, "Initializing integration cache...")

	enabled := true
	views, _, err := s.readRepo.List(ctx, domain.IntegrationFilter{
		Enabled: &enabled,
		Limit:   10000,
	})
	if err != nil {
		return fmt.Errorf("integration_cache.init: list integrations: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.integrations = make(map[domain.IntegrationID]*CachedIntegration, len(views))
	s.apiKeys = make(map[string]*CachedIntegration, len(views))

	for _, v := range views {
		ci := &CachedIntegration{
			ID:         v.ID,
			Name:       v.Name,
			Type:       v.Type,
			APIKey:     v.APIKey,
			WebhookURL: v.WebhookURL,
			Enabled:    v.Enabled,
			Config:     v.Config,
		}
		s.integrations[v.ID] = ci
		if v.APIKey != "" {
			s.apiKeys[v.APIKey] = ci
		}
	}

	s.logger.Infoc(ctx, "Integration cache initialized",
		"integrations_count", len(s.integrations),
		"api_keys_count", len(s.apiKeys),
	)
	return nil
}

// InvalidateCache refreshes the cache when a database change is detected.
func (s *CacheService) InvalidateCache(ctx context.Context, table string) error {
	if table == consts.TableIntegrations || table == consts.TableAPIKeys {
		s.logger.Infoc(ctx, "Invalidating integration cache", "table", table)
		return s.InitCache(ctx)
	}
	return nil
}

// FindByAPIKey returns a cached integration by its API key.
func (s *CacheService) FindByAPIKey(key string) (*CachedIntegration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ci, ok := s.apiKeys[key]
	return ci, ok
}

// FindByID returns a cached integration by ID.
func (s *CacheService) FindByID(id domain.IntegrationID) (*CachedIntegration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ci, ok := s.integrations[id]
	return ci, ok
}
