package ratelimit

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateRateLimitRequest) (*domain.RateLimit, error) {
	rl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		rl.Name = *req.Name
	}
	if req.PathPattern != nil {
		rl.PathPattern = *req.PathPattern
	}
	if req.Method != nil {
		rl.Method = *req.Method
	}
	if req.LimitCount != nil {
		rl.LimitCount = *req.LimitCount
	}
	if req.WindowSeconds != nil {
		rl.WindowSeconds = *req.WindowSeconds
	}
	if req.IsActive != nil {
		rl.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, rl); err != nil {
		return nil, err
	}
	return rl, nil
}
