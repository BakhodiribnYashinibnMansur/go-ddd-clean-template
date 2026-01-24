package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error) {
	qb := r.builder.
		Select("id, role_id, username, email, phone, password_hash, salt, attributes, active, is_approved, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.RoleID != nil {
		qb = qb.Where(squirrel.Eq{"role_id": *filter.RoleID})
	}

	if filter.Username != nil {
		qb = qb.Where(squirrel.Eq{"username": *filter.Username})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	if filter.Email != nil {
		qb = qb.Where(squirrel.Eq{"email": *filter.Email})
	}

	if filter.Active != nil {
		qb = qb.Where(squirrel.Eq{"active": *filter.Active})
	}

	if filter.IsApproved != nil {
		qb = qb.Where(squirrel.Eq{"is_approved": *filter.IsApproved})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build SQL query")
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.RoleID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Salt,
		&u.Attributes, &u.Active, &u.IsApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"filter": filter,
		})
	}

	return &u, nil
}
