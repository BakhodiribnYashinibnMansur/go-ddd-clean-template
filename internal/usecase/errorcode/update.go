package errorcode

import (
	"context"
	"gct/internal/domain"
	repo "gct/internal/repo/persistent/postgres/errorcode"
)

func (uc *UseCase) Update(ctx context.Context, code string, input repo.UpdateErrorCodeInput) (*domain.ErrorCode, error) {
	return uc.repo.Update(ctx, code, input)
}
