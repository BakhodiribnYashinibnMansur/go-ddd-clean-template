package iprule

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.IPRuleFilter) ([]domain.IPRule, int64, error) {
	return uc.repo.List(ctx, filter)
}
