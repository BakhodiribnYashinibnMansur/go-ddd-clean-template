package role

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

type UseCaseI interface {
	Create(ctx context.Context, role *domain.Role) error
	Get(ctx context.Context, filter *domain.RoleFilter) (*domain.Role, error)
	Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	Assign(ctx context.Context, userID, roleID uuid.UUID) error

	// Permission management
	AddPermission(ctx context.Context, roleID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error
}
