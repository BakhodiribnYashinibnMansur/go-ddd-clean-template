package client

import (
	"context"

	"github.com/Masterminds/squirrel"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error) {
	qb := r.builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	if !filter.IsIDNull() {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if !filter.IsPhoneNull() {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build SQL query")
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"filter_id":    filter.ID,
			"filter_phone": filter.Phone,
		})
	}

	return &u, nil
}
