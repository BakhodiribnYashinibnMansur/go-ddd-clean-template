package user

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
)

// UseCase -.
type UseCase struct {
	repo repo.UserRepo
}

// New -.
func New(r repo.UserRepo) *UseCase {
	return &UseCase{
		repo: r,
	}
}

func (uc *UseCase) Create(ctx context.Context, u entity.User) error {
	return uc.repo.Create(ctx, u)
}

func (uc *UseCase) GetByID(ctx context.Context, id int64) (entity.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UseCase) GetByPhone(ctx context.Context, phone string) (entity.User, error) {
	return uc.repo.GetByPhone(ctx, phone)
}

func (uc *UseCase) Update(ctx context.Context, u entity.User) error {
	userData, err := uc.repo.GetByID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("UserUseCase - Update - uc.repo.GetByID: %w", err)
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

	return uc.repo.Update(ctx, userData)
}

func (uc *UseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
