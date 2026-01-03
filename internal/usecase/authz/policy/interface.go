package policy

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, policy *domain.Policy) error
	Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error)
	Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error)
	Update(ctx context.Context, policy *domain.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
}
