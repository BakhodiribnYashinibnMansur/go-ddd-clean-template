package client

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error) {
	qb := r.buildSelectUsersQuery(filter)
	countQb := r.buildCountUsersQuery(filter)

	count, err := r.getTotalCount(ctx, countQb)
	if err != nil {
		return nil, 0, err
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select SQL query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "users", map[string]any{"operation": "get_users"})
	}
	defer rows.Close()

	users, err := r.scanUserRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (r *Repo) buildSelectUsersQuery(filter *domain.UsersFilter) squirrel.SelectBuilder {
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

	if filter.IsValidLimit() {
		qb = qb.Limit(uint64(filter.Pagination.Limit))
	}
	if filter.IsValidOffset() {
		qb = qb.Offset(uint64(filter.Pagination.Offset))
	}

	return qb
}

func (r *Repo) buildCountUsersQuery(filter *domain.UsersFilter) squirrel.SelectBuilder {
	countQb := r.builder.Select("COUNT(*)").From("users").Where("deleted_at = 0")
	if filter.ID != nil {
		countQb = countQb.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Phone != nil {
		countQb = countQb.Where(squirrel.Eq{"phone": *filter.Phone})
	}
	return countQb
}

func (r *Repo) getTotalCount(ctx context.Context, qb squirrel.SelectBuilder) (int, error) {
	sql, args, err := qb.ToSql()
	if err != nil {
		return 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count SQL query")
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, apperrors.HandlePgError(ctx, err, "users", map[string]any{"operation": "count"})
	}
	return count, nil
}

func (r *Repo) scanUserRows(ctx context.Context, rows pgx.Rows) ([]*domain.User, error) {
	var users []*domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
		)
		if err != nil {
			return nil, apperrors.HandlePgError(ctx, err, "users", map[string]any{"operation": "scan_row"})
		}
		users = append(users, &u)
	}
	return users, nil
}
