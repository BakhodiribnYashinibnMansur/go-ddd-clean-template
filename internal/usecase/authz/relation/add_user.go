package relation

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) AddUser(ctx context.Context, userID, relationID uuid.UUID) error {
	u.logger.Infow("relation add user started", "user_id", userID, "relation_id", relationID)

	err := u.repo.Postgres.Authz.Relation.AddUser(ctx, relationID, userID)
	if err != nil {
		u.logger.Errorw("relation add user failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"userID": userID, "relationID": relationID})
	}
	u.logger.Infow("relation add user success")
	return nil
}
