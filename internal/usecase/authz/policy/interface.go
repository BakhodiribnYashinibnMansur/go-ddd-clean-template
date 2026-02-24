package policy

import (
	"context"

	"gct/internal/domain"
	"github.com/google/uuid"
)

type UseCaseI interface {
	Create(ctx context.Context, policy *domain.Policy) error
	Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error)
	Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error)
	Update(ctx context.Context, policy *domain.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	Toggle(ctx context.Context, id uuid.UUID) error
}
