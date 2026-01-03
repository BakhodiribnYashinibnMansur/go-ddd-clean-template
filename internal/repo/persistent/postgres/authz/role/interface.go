package role

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type RepoI interface {
	Create(ctx context.Context, r *domain.Role) error
	Get(ctx context.Context, filter *domain.RoleFilter) (*domain.Role, error)
	Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error)
	Update(ctx context.Context, r *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddPermission(ctx context.Context, roleID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
}
