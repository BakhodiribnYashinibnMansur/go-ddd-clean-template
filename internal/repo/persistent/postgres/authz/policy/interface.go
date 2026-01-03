package policy

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type RepoI interface {
	Create(ctx context.Context, p *domain.Policy) error
	Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error)
	Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error)
	Update(ctx context.Context, p *domain.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]*domain.Policy, error)
}
