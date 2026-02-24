package announcement

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Toggle(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	a, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.IsActive = !a.IsActive
	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
