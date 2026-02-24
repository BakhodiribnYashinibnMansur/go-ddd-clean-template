package ratelimit

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateRateLimitRequest) (*domain.RateLimit, error) {
	rl := &domain.RateLimit{
		ID:            uuid.New(),
		Name:          req.Name,
		PathPattern:   req.PathPattern,
		Method:        req.Method,
		LimitCount:    req.LimitCount,
		WindowSeconds: req.WindowSeconds,
		IsActive:      req.IsActive,
	}
	if err := uc.repo.Create(ctx, rl); err != nil {
		return nil, err
	}
	return rl, nil
}
