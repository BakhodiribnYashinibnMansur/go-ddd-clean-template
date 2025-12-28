package client

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error) {
	qb := r.builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build SQL query")).
			WithField("operation", "build_query").
			WithDetails("Error occurred while building SQL query with squirrel")
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.AutoSource(
				apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound,
					"user not found in database")).
				WithField("table", "users").
				WithField("filter_id", filter.ID).
				WithField("filter_phone", filter.Phone).
				WithDetails("No user record exists with the given filter criteria")
		}

		return nil, apperrors.AutoSource(
			apperrors.WrapRepoError(ctx, err, apperrors.ErrRepoDatabase,
				"failed to query user from database")).
			WithField("table", "users").
			WithField("filter_id", filter.ID).
			WithField("filter_phone", filter.Phone)
	}

	return &u, nil
}

func (r *Repo) User(ctx context.Context, id int64) (*domain.User, error) {
	return r.Get(ctx, &domain.UserFilter{ID: &id})
}
