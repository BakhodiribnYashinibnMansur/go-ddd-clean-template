package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (uc *UseCase) Create(ctx context.Context, u domain.User) error {
	return uc.repo.User.Client.Create(ctx, u)
}
