package job

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateJobRequest) (*domain.Job, error) {
	j, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		j.Name = *req.Name
	}
	if req.Type != nil {
		j.Type = *req.Type
	}
	if req.CronSchedule != nil {
		j.CronSchedule = *req.CronSchedule
	}
	if req.Payload != nil {
		j.Payload = req.Payload
	}
	if req.IsActive != nil {
		j.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}
