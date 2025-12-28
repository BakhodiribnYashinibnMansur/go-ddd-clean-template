package client

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

func (uc *UseCase) SignOut(ctx context.Context, in SignOutInput) error {
	sessionID, err := uuid.Parse(in.SessionID)
	if err != nil {
		return err
	}

	return uc.repo.User.SessionRepo.Revoke(ctx, &domain.SessionFilter{ID: sessionID})
}
