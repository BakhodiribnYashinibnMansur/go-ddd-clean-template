package systemerror

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.SystemError) error
	Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error)
	Resolve(ctx context.Context, id string, resolvedBy *string) error
}
