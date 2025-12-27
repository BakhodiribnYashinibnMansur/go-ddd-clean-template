package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (r *Repo) GetByPhone(ctx context.Context, phone string) (domain.User, error) {
	return r.Get(ctx, UserFilter{Phone: &phone})
}
