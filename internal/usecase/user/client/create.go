package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	uc.logger.Infow("user create started", "input", u)

	err := uc.repo.Postgres.User.Client.Create(ctx, u)
	if err != nil {
		uc.logger.Errorw("user create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
	}

	uc.logger.Infow("user create success")
	return nil
}
