package client

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error) {
	qb := r.builder.
		Select(
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
		From(tableName).
		Where(schema.UsersDeletedAt + " = 0")

	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersID: *filter.ID})
	}

	if filter.RoleID != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersRoleID: *filter.RoleID})
	}

	if filter.Username != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersUsername: *filter.Username})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersPhone: *filter.Phone})
	}

	if filter.Email != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersEmail: *filter.Email})
	}

	if filter.Active != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersActive: *filter.Active})
	}

	if filter.IsApproved != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersIsApproved: *filter.IsApproved})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildQuery)
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.RoleID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Salt,
		&u.Attributes, &u.Active, &u.IsApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{
			"filter": filter,
		})
	}

	return &u, nil
}
