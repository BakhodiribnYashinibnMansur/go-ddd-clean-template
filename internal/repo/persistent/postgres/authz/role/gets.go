package role

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.RolesFilter) ([]*domain.Role, int, error) {
	query := r.builder.Select("id", "name", "created_at").From("role")
	countQuery := r.builder.Select("COUNT(*)").From("role")

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Name != nil {
		query = query.Where(squirrel.Eq{"name": *filter.Name})
		countQuery = countQuery.Where(squirrel.Eq{"name": *filter.Name})
	}

	if filter.Pagination != nil {
		query = query.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "role", nil)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt); err != nil {
			return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %v", err))
		}
		roles = append(roles, &role)
	}

	// Count
	var count int
	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count query")
	}
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "role", nil)
	}

	return roles, count, nil
}
