package client

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (r *Repo) Users(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error) {
	// Base query
	qb := r.builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	// Apply filters
	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	// Count query
	countQb := r.builder.Select("COUNT(*)").From("users").Where("deleted_at = 0")
	if filter.ID != nil {
		countQb = countQb.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Phone != nil {
		countQb = countQb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	countSql, countArgs, err := countQb.ToSql()
	if err != nil {
		return nil, 0, apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build count SQL query")).
			WithDetails("Error occurred while building COUNT query for users")
	}

	var count int
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"operation": "count",
		})
	}

	// Apply pagination
	if filter.Limit > 0 {
		qb = qb.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		qb = qb.Offset(uint64(filter.Offset))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build select SQL query")).
			WithDetails("Error occurred while building SELECT query for users")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "users", map[string]any{
			"operation": "get_users",
			"limit":     filter.Limit,
			"offset":    filter.Offset,
		})
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		err = rows.Scan(
			&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
		)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(ctx, err, "users", map[string]any{
				"operation": "scan_row",
			})
		}
		users = append(users, &u)
	}

	return users, count, nil
}
