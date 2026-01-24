package relation

import (
	"context"

	apperrors "gct/pkg/errors"
	"github.com/google/uuid"
)

func (u *UseCase) RemoveUser(ctx context.Context, userID, relationID uuid.UUID) error {
	u.logger.WithContext(ctx).Infow("relation remove user started", "user_id", userID, "relation_id", relationID)

	err := u.repo.Postgres.Authz.Relation.RemoveUser(ctx, relationID, userID)
	if err != nil {
		u.logger.WithContext(ctx).Errorw("relation remove user failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(map[string]any{"userID": userID, "relationID": relationID})
	}
	u.logger.WithContext(ctx).Infow("relation remove user success")
	return nil
}
