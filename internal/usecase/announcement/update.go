package announcement

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateAnnouncementRequest) (*domain.Announcement, error) {
	a, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Content != nil {
		a.Content = *req.Content
	}
	if req.Type != nil {
		a.Type = *req.Type
	}
	if req.IsActive != nil {
		a.IsActive = *req.IsActive
	}
	if req.StartsAt != nil {
		a.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		a.EndsAt = req.EndsAt
	}
	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
