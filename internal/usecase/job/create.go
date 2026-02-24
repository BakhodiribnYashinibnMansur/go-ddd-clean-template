package job

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateJobRequest) (*domain.Job, error) {
	j := &domain.Job{
		ID:           uuid.New(),
		Name:         req.Name,
		Type:         req.Type,
		CronSchedule: req.CronSchedule,
		Payload:      req.Payload,
		IsActive:     req.IsActive,
		Status:       "idle",
	}
	if j.Payload == nil {
		j.Payload = map[string]any{}
	}
	if err := uc.repo.Create(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}
