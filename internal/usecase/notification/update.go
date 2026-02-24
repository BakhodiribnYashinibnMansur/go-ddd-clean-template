package notification

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateNotificationRequest) (*domain.Notification, error) {
	n, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		n.Title = *req.Title
	}
	if req.Body != nil {
		n.Body = *req.Body
	}
	if req.Type != nil {
		n.Type = *req.Type
	}
	if req.TargetType != nil {
		n.TargetType = *req.TargetType
	}
	if req.IsActive != nil {
		n.IsActive = *req.IsActive
	}
	if err := uc.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}
