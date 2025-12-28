package session

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

// Revoke revokes a session.
func (uc *UseCase) Revoke(ctx context.Context, filter *domain.SessionFilter) error {
	return uc.repo.User.SessionRepo.Revoke(ctx, filter)
}
