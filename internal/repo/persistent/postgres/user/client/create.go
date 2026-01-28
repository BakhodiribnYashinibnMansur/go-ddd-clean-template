package client

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, u *domain.User) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			schema.UsersID,
			schema.UsersRoleID,
			schema.UsersUsername,
			schema.UsersEmail,
			schema.UsersPhone,
			schema.UsersPasswordHash,
			schema.UsersSalt,
			schema.UsersAttributes,
			schema.UsersActive,
			schema.UsersIsApproved,
			schema.UsersCreatedAt,
			schema.UsersUpdatedAt,
			schema.UsersDeletedAt,
			schema.UsersLastSeen,
		).
		Values(u.ID, u.RoleID, u.Username, u.Email, u.Phone, u.PasswordHash, u.Salt, u.Attributes, u.Active, u.IsApproved, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildInsert)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
