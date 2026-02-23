package translation

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) Delete(ctx context.Context, filter domain.TranslationFilter) error {
	return uc.repo.Delete(ctx, filter)
}
