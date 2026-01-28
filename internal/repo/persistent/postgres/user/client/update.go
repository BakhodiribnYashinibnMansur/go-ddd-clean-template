package client

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Update(ctx context.Context, u *domain.User) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set(schema.UsersRoleID, u.RoleID).
		Set(schema.UsersUsername, u.Username).
		Set(schema.UsersEmail, u.Email).
		Set(schema.UsersPhone, u.Phone).
		Set(schema.UsersPasswordHash, u.PasswordHash).
		Set(schema.UsersSalt, u.Salt).
		Set(schema.UsersAttributes, u.Attributes).
		Set(schema.UsersActive, u.Active).
		Set(schema.UsersIsApproved, u.IsApproved).
		Set(schema.UsersUpdatedAt, time.Now()).
		Set(schema.UsersLastSeen, u.LastSeen).
		Where(schema.UsersID+" = ? AND "+schema.UsersDeletedAt+" = 0", u.ID).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildUpdate)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
