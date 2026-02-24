package job

import (
	"context"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Trigger(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	j, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	j.Status = "running"
	j.LastRunAt = &now
	if err := uc.repo.Update(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}
