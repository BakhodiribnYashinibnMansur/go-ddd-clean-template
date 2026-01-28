package errorcode

import (
	"context"
	"gct/internal/domain"
	repo "gct/internal/repo/persistent/postgres/errorcode"
)

func (uc *UseCase) Create(ctx context.Context, input repo.CreateErrorCodeInput) (*domain.ErrorCode, error) {
	return uc.repo.Create(ctx, input)
}
