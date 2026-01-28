package errorcode

import (
	"context"
	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context) ([]*domain.ErrorCode, error) {
	return uc.repo.List(ctx)
}
