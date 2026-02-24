package webhook

import (
	"context"

	"github.com/google/uuid"
)

func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}
