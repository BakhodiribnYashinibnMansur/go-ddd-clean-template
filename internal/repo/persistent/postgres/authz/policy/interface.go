package policy

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

type RepoI interface {
	Create(ctx context.Context, p *domain.Policy) error
	Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error)
	Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error)
	Update(ctx context.Context, p *domain.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]*domain.Policy, error)
	Toggle(ctx context.Context, id uuid.UUID) error
}
