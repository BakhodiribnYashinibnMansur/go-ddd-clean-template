package ratelimit

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Toggle(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error) {
	rl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rl.IsActive = !rl.IsActive
	if err := uc.repo.Update(ctx, rl); err != nil {
		return nil, err
	}
	return rl, nil
}
