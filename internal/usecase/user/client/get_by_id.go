package client

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
)

// GetByID gets a user by ID.
func (uc *UseCase) GetByID(ctx context.Context, id int64) (domain.User, error) {
	user, err := uc.repo.User.Client.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("ClientUseCase - GetByID - uc.repo.User.Client.GetByID: %w", err)
	}
	return user, nil
}
