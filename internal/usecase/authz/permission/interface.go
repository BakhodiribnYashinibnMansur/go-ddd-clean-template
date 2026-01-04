package permission

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

type UseCaseI interface {
	Create(ctx context.Context, perm *domain.Permission) error
	Get(ctx context.Context, filter *domain.PermissionFilter) (*domain.Permission, error)
	Gets(ctx context.Context, filter *domain.PermissionsFilter) ([]*domain.Permission, int, error)
	Update(ctx context.Context, perm *domain.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Scope Management
	AssignScope(ctx context.Context, permID uuid.UUID, path, method string) error
	RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error

	// Role assignment helper (alias)
	AssignToRole(ctx context.Context, roleID, permID uuid.UUID) error
}
