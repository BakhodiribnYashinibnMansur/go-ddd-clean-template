package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (uc *UseCase) GetByPhone(ctx context.Context, phone string) (domain.User, error) {
	return uc.repo.User.Client.GetByPhone(ctx, phone)
}
