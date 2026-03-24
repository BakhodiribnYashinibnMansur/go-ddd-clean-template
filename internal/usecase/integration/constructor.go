package integration

import (
	"sync"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

type UseCase struct {
	repo   Repository
	logger logger.Log
	cfg    *config.Config

	// In-memory cache for fast validation
	mu           sync.RWMutex
	integrations map[uuid.UUID]*domain.Integration
	apiKeys      map[string]*domain.APIKey // Key is hashed key string
}

func New(repo Repository, logger logger.Log, cfg *config.Config) UseCaseI {
	return &UseCase{
		repo:         repo,
		logger:       logger,
		cfg:          cfg,
		integrations: make(map[uuid.UUID]*domain.Integration),
		apiKeys:      make(map[string]*domain.APIKey),
	}
}
