package scope

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, scope *domain.Scope) error
	Get(ctx context.Context, filter *domain.ScopeFilter) (*domain.Scope, error)
	Gets(ctx context.Context, filter *domain.ScopesFilter) ([]*domain.Scope, int, error)
	Delete(ctx context.Context, path, method string) error
}
