package announcement

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateAnnouncementRequest) (*domain.Announcement, error) {
	a := &domain.Announcement{
		ID:       uuid.New(),
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		IsActive: req.IsActive,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
	}
	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
