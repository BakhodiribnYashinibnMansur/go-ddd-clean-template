package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (r *Repo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	return r.Get(ctx, UserFilter{ID: &id})
}
