package integration

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// Repository handles integration-related database operations.
type Repository interface {
	// Integration CRUD
	CreateIntegration(ctx context.Context, integration *domain.Integration) error
	GetIntegrationByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error)
	GetIntegrationByName(ctx context.Context, name string) (*domain.Integration, error)
	ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error)
	UpdateIntegration(ctx context.Context, integration *domain.Integration) error
	DeleteIntegration(ctx context.Context, id uuid.UUID) error

	// API Key CRUD
	CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error
	GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	GetAPIKeyByKey(ctx context.Context, key string) (*domain.APIKey, error)
	ListAPIKeysByIntegration(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *domain.APIKey) error
	DeleteAPIKey(ctx context.Context, id uuid.UUID) error
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error
}

// UseCaseI defines the business logic interface for integrations.
type UseCaseI interface {
	// Integration operations
	CreateIntegration(ctx context.Context, req domain.CreateIntegrationRequest) (*domain.Integration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*domain.IntegrationWithKeys, error)
	ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error)
	UpdateIntegration(ctx context.Context, id uuid.UUID, req domain.UpdateIntegrationRequest) (*domain.Integration, error)
	DeleteIntegration(ctx context.Context, id uuid.UUID) error
	ToggleIntegration(ctx context.Context, id uuid.UUID) (*domain.Integration, error)

	// API Key operations
	CreateAPIKey(ctx context.Context, req domain.CreateAPIKeyRequest) (*domain.APIKey, string, error) // Returns APIKey and raw key
	GetAPIKey(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	ListAPIKeys(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error)
	ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error)
	RevokeAPIKey(ctx context.Context, id uuid.UUID) error
	DeleteAPIKey(ctx context.Context, id uuid.UUID) error

	// Reactive Caching
	InitCache(ctx context.Context) error
	InvalidateCache(ctx context.Context, table string) error
}
