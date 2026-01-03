package systemError

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.SystemError) error
	Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error)
}
