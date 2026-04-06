package repository

import (
	"context"

	"gct/internal/context/admin/supporting/integration/domain/entity"
)

// IntegrationRepository is the write-side repository for the Integration aggregate.
// Delete performs a hard delete — callers should ensure authorization before invoking.
type IntegrationRepository interface {
	Save(ctx context.Context, e *entity.Integration) error
	FindByID(ctx context.Context, id entity.IntegrationID) (*entity.Integration, error)
	Update(ctx context.Context, e *entity.Integration) error
	Delete(ctx context.Context, id entity.IntegrationID) error
}
