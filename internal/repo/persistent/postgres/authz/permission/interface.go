package permission

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type RepoI interface {
	Create(ctx context.Context, p *domain.Permission) error
	Get(ctx context.Context, filter *domain.PermissionFilter) (*domain.Permission, error)
	Gets(ctx context.Context, filter *domain.PermissionsFilter) ([]*domain.Permission, int, error)
	Update(ctx context.Context, p *domain.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddScope(ctx context.Context, permID uuid.UUID, path, method string) error
	RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error
	GetScopes(ctx context.Context, permID uuid.UUID) ([]*domain.Scope, error)
}
