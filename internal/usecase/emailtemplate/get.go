package emailtemplate

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	return uc.repo.GetByID(ctx, id)
}
