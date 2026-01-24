package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error) {
	// Base query
	baseQb := r.builder.Select("id, role_id, username, email, phone, password_hash, salt, attributes, active, is_approved, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	// Apply dynamic filters
	qb := r.applyFilters(baseQb, &filter.UserFilter)

	// Apply Sorting & Pagination
	if filter.Pagination != nil {
		if filter.Pagination.SortBy != "" {
			order := "ASC"
			if filter.Pagination.SortOrder == "DESC" {
				order = "DESC"
			}
			qb = qb.OrderBy(filter.Pagination.SortBy + " " + order)
		} else {
			qb = qb.OrderBy("created_at DESC")
		}
		qb = qb.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build SQL query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "users", nil)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.RoleID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Salt,
			&u.Attributes, &u.Active, &u.IsApproved,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
		); err != nil {
			return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to scan row")
		}
		users = append(users, &u)
	}

	// Count Query
	// Start fresh from builder but reuse filter logic
	countBaseQb := r.builder.Select("COUNT(*)").From("users").Where("deleted_at = 0")
	countQb := r.applyFilters(countBaseQb, &filter.UserFilter)

	countSql, countArgs, err := countQb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count query")
	}
	var count int
	if err := r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count); err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "users", nil)
	}

	return users, count, nil
}

func (r *Repo) applyFilters(qb squirrel.SelectBuilder, filter *domain.UserFilter) squirrel.SelectBuilder {
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
	return qb
}
