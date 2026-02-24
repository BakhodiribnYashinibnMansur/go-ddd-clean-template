package announcement

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	return uc.repo.GetByID(ctx, id)
}
