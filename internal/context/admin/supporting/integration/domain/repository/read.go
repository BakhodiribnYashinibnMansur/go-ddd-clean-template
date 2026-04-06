package repository

import (
	"context"

	"gct/internal/context/admin/supporting/integration/domain/entity"
)

// IntegrationReadRepository is the read-side repository returning projected views.
// Implementations must return ErrIntegrationNotFound when FindByID yields no result.
type IntegrationReadRepository interface {
	FindByID(ctx context.Context, id entity.IntegrationID) (*entity.IntegrationView, error)
	List(ctx context.Context, filter entity.IntegrationFilter) ([]*entity.IntegrationView, int64, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*entity.IntegrationAPIKeyView, error)

	// ListActiveJWT returns all integrations that have jwt_api_key_hash set
	// (NULL means JWT is not provisioned for that integration yet).
	ListActiveJWT(ctx context.Context) ([]entity.JWTIntegrationView, error)

	// FindJWTByHash returns the integration whose jwt_api_key_hash exactly
	// matches the provided hash. Uses the DB unique index. Returns
	// ErrIntegrationNotFound if not found.
	FindJWTByHash(ctx context.Context, hash []byte) (*entity.JWTIntegrationView, error)
}
