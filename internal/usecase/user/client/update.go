package client

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
)

func (uc *UseCase) Update(ctx context.Context, u domain.User) error {
	userData, err := uc.repo.User.Client.GetByID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("UserUseCase - Update - uc.repo.User.Client.GetByID: %w", err)
	}

	// Update fields if provided
	if u.Username != nil {
		userData.Username = u.Username
	}
	if u.Phone != "" {
		userData.Phone = u.Phone
	}
	if u.PasswordHash != "" {
		userData.PasswordHash = u.PasswordHash
	}
	if u.Salt != nil {
		userData.Salt = u.Salt
	}
	if u.LastSeen != nil {
		userData.LastSeen = u.LastSeen
	}

	return uc.repo.User.Client.Update(ctx, userData)
}
