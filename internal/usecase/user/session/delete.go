package session

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

// Delete terminates a session.
func (uc *UseCase) Delete(ctx context.Context, filter *domain.SessionFilter) error {
	return uc.repo.User.SessionRepo.Delete(ctx, filter)
}
