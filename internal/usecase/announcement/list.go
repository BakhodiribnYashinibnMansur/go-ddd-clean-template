package announcement

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.AnnouncementFilter) ([]domain.Announcement, int64, error) {
	return uc.repo.List(ctx, filter)
}
