package relation

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) RemoveUser(ctx context.Context, userID, relationID uuid.UUID) error {
	u.logger.Infow("relation remove user started", "user_id", userID, "relation_id", relationID)

	err := u.repo.Postgres.Authz.Relation.RemoveUser(ctx, relationID, userID)
	if err != nil {
		u.logger.Errorw("relation remove user failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(map[string]any{"userID": userID, "relationID": relationID})
	}
	u.logger.Infow("relation remove user success")
	return nil
}
