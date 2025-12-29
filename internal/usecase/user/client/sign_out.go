package client

import (
	"context"

	"github.com/google/uuid"

	"gct/internal/domain"
)

func (uc *UseCase) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	sessionID, err := uuid.Parse(in.SessionID)
	if err != nil {
		return err
	}

	return uc.repo.Postgres.SessionRepo.Revoke(ctx, &domain.SessionFilter{ID: &sessionID})
}
