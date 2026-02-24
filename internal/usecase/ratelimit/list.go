package ratelimit

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.RateLimitFilter) ([]domain.RateLimit, int64, error) {
	return uc.repo.List(ctx, filter)
}
