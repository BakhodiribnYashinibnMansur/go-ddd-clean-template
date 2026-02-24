package dashboard

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) Get(ctx context.Context) (domain.DashboardStats, error) {
	return uc.repo.Get(ctx)
}
