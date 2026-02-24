package notification

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateNotificationRequest) (*domain.Notification, error) {
	n := &domain.Notification{
		ID:         uuid.New(),
		Title:      req.Title,
		Body:       req.Body,
		Type:       req.Type,
		TargetType: req.TargetType,
		IsActive:   req.IsActive,
	}
	if err := uc.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}
