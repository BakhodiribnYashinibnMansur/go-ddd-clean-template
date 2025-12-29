package client

import (
	"context"

	"gct/internal/domain"
)

func (r *Repo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return r.Get(ctx, &domain.UserFilter{Phone: &phone})
}
