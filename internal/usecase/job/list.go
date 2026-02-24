package job

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.JobFilter) ([]domain.Job, int64, error) {
	return uc.repo.List(ctx, filter)
}
