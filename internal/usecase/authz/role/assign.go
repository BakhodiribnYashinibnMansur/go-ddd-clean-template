package role

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (u *UseCase) Assign(ctx context.Context, userID, roleID uuid.UUID) error {
	u.logger.Infoc(ctx, "role assign started", "user_id", userID, "role_id", roleID)

	user, err := u.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &userID})
	if err != nil {
		u.logger.Errorc(ctx, "role assign failed: get user", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(userID)
	}

	user.RoleID = &roleID
	err = u.repo.Postgres.User.Client.Update(ctx, user)
	if err != nil {
		u.logger.Errorc(ctx, "role assign failed: update user", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(map[string]any{"userID": userID, "roleID": roleID})
	}

	u.logger.Infoc(ctx, "role assign success")
	return nil
}
