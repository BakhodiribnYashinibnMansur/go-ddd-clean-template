package errorcode

import (
	"context"
	"gct/internal/domain"
)

func (uc *UseCase) GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error) {
	return uc.repo.GetByCode(ctx, code)
}
